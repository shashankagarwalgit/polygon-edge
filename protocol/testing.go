package protocol

import (
	"context"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-sdk/blockchain"
	"github.com/0xPolygon/polygon-sdk/helper/tests"
	"github.com/0xPolygon/polygon-sdk/network"
	"github.com/0xPolygon/polygon-sdk/types"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

const (
	maxHeight = 1024
	maxSeed   = math.MaxInt32
)

var (
	defaultNetworkConfig = func(c *network.Config) {
		c.NoDiscover = true
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func CreateSyncer(t *testing.T, blockchain blockchainShim, serverCfg *func(c *network.Config)) *Syncer {
	t.Helper()
	if serverCfg == nil {
		serverCfg = &defaultNetworkConfig
	}

	srv := network.CreateServer(t, *serverCfg)
	syncer := NewSyncer(hclog.NewNullLogger(), srv, blockchain)
	syncer.Start()

	return syncer
}

func WaitUntilPeerConnected(t *testing.T, syncer *Syncer, numPeer int, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(func() {
		cancel()
	})

	_, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		num := len(syncer.peers)
		if num == numPeer {
			return nil, false
		}
		return nil, true
	})
	assert.NoError(t, err)
}

func WaitUntilProcessedAllEvents(t *testing.T, syncer *Syncer, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(func() {
		cancel()
	})

	_, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		return nil, len(syncer.blockchain.SubscribeEvents().GetEventCh()) > 0
	})
	assert.NoError(t, err)
}

func NewRandomChain(t *testing.T, height int) blockchainShim {
	seed := rand.Intn(maxSeed)
	return blockchain.NewTestBlockchain(t, blockchain.NewTestHeaderChainWithSeed(nil, height, seed))
}

func SetupSyncerNetwork(t *testing.T, chain blockchainShim, peerChains []blockchainShim) (syncer *Syncer, peerSyncers []*Syncer) {
	t.Helper()

	syncer = CreateSyncer(t, chain, nil)
	peerSyncers = make([]*Syncer, len(peerChains))
	for idx, peerChain := range peerChains {
		peerSyncers[idx] = CreateSyncer(t, peerChain, nil)
		network.MultiJoin(t, syncer.server, peerSyncers[idx].server)
	}
	WaitUntilPeerConnected(t, syncer, len(peerChains), 10*time.Second)
	return
}

func GenerateNewBlocks(t *testing.T, chain blockchainShim, num int) []*types.Block {
	t.Helper()

	currentHeight := chain.Header().Number
	oldHeaders := make([]*types.Header, currentHeight+1)
	for i := uint64(1); i <= currentHeight; i++ {
		var ok bool
		oldHeaders[i], ok = chain.GetHeaderByNumber(i)
		assert.Truef(t, ok, "chain should have header at %d, but empty", i)
	}
	headers := blockchain.NewTestHeaderFromChain(oldHeaders, num)
	return blockchain.HeadersToBlocks(headers[currentHeight+1:])
}

func TryPopBlock(t *testing.T, syncer *Syncer, peerID peer.ID, timeout time.Duration) (*types.Block, bool) {
	t.Helper()

	peer := syncer.peers[peerID]
	assert.NotNil(t, peer, "syncer doesn't have peer %s", peerID.String())

	blockCh := make(chan *types.Block, 1)
	go func() {
		blockCh <- peer.popBlock()
	}()

	select {
	case block := <-blockCh:
		return block, true
	case <-time.After(timeout):
		return nil, false
	}
}

func GetCurrentStatus(b blockchainShim) *Status {
	return &Status{
		Hash:       b.Header().Hash,
		Number:     b.Header().Number,
		Difficulty: b.CurrentTD(),
	}
}

func HeaderToStatus(h *types.Header) *Status {
	return &Status{
		Hash:       h.Hash,
		Number:     h.Number,
		Difficulty: big.NewInt(0).SetUint64(h.Difficulty),
	}
}

func BlockToStatus(b *types.Block) *Status {
	return HeaderToStatus(b.Header)
}

type mockBlockchain struct {
	blocks       []*types.Block
	subscription *mockSubscription
}

func NewMockBlockchain(headers []*types.Header) *mockBlockchain {
	return &mockBlockchain{
		blocks:       blockchain.HeadersToBlocks(headers),
		subscription: NewMockSubscription(),
	}
}

func (b *mockBlockchain) SubscribeEvents() blockchain.Subscription {
	return b.subscription
}

func (b *mockBlockchain) Header() *types.Header {
	l := len(b.blocks)
	if l == 0 {
		return nil
	}
	return b.blocks[l-1].Header
}

func (b *mockBlockchain) CurrentTD() *big.Int {
	current := b.Header()
	if current == nil {
		return nil
	}
	return new(big.Int).SetUint64(current.Difficulty)
}

func (b *mockBlockchain) GetTD(hash types.Hash) (*big.Int, bool) {
	for _, b := range b.blocks {
		if b.Header.Hash == hash {
			return new(big.Int).SetUint64(b.Header.Difficulty), true
		}
	}
	return nil, false
}

func (b *mockBlockchain) GetReceiptsByHash(types.Hash) ([]*types.Receipt, error) {
	panic("not implement")
}

func (b *mockBlockchain) GetBodyByHash(types.Hash) (*types.Body, bool) {
	panic("not implement")
}

func (b *mockBlockchain) GetHeaderByHash(h types.Hash) (*types.Header, bool) {
	for _, b := range b.blocks {
		if b.Header.Hash == h {
			return b.Header, true
		}
	}
	return nil, false
}

func (b *mockBlockchain) GetHeaderByNumber(n uint64) (*types.Header, bool) {
	for _, b := range b.blocks {
		if b.Header.Number == n {
			return b.Header, true
		}
	}
	return nil, false
}

func (b *mockBlockchain) WriteBlocks(blocks []*types.Block) error {
	b.blocks = append(b.blocks, blocks...)
	b.subscription.AppendBlocks(blocks)
	return nil
}

type mockSubscription struct {
	eventCh chan *blockchain.Event
}

func NewMockSubscription() *mockSubscription {
	return &mockSubscription{
		eventCh: make(chan *blockchain.Event, 2), // make with 2 capacities in order to check subsequent event easily
	}
}

func (s *mockSubscription) AppendBlocks(blocks []*types.Block) {
	for _, b := range blocks {
		s.eventCh <- &blockchain.Event{
			Difficulty: new(big.Int).SetUint64(b.Header.Difficulty),
			NewChain:   []*types.Header{b.Header},
		}
	}
}

func (s *mockSubscription) GetEventCh() chan *blockchain.Event {
	return s.eventCh
}

func (s *mockSubscription) GetEvent() *blockchain.Event {
	return <-s.eventCh
}

func (s *mockSubscription) Close() {
	close(s.eventCh)
}
