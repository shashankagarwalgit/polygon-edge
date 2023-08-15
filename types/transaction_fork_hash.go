package types

import (
	"math/big"

	"github.com/0xPolygon/polygon-edge/forkmanager"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/umbracle/fastrlp"
)

const txHashHandler = "txHash"

type TransactionHashFork interface {
	SerializeForRootCalculation(*Transaction, *fastrlp.ArenaPool) []byte
	ComputeHash(*Transaction)
}

var (
	_ TransactionHashFork = (*TransactionHashForkV1)(nil)
	_ TransactionHashFork = (*TransactionHashForkV2)(nil)
)

type TransactionHashForkV1 struct {
}

func (th *TransactionHashForkV1) SerializeForRootCalculation(t *Transaction, ap *fastrlp.ArenaPool) []byte {
	ar := ap.Get()
	chainID := t.ChainID()
	//t.ChainID = big.NewInt(0)
	t.SetChainID(big.NewInt(0))

	defer func() {
		ap.Put(ar)

		//t.ChainID = chainID
		t.SetChainID(chainID)
	}()

	ar.Reset()

	return t.MarshalRLPWith(ar).MarshalTo(nil)
}

func (th *TransactionHashForkV1) ComputeHash(t *Transaction) {
	ar := marshalArenaPool.Get()
	hash := keccak.DefaultKeccakPool.Get()

	chainID := t.ChainID()
	// t.ChainID = big.NewInt(0)
	t.SetChainID(big.NewInt(0))

	v := t.MarshalRLPWith(ar)

	var txHash Hash

	hash.WriteRlp(txHash[:0], v)
	t.SetHash(txHash)

	// t.ChainID = chainID
	t.SetChainID(chainID)

	marshalArenaPool.Put(ar)
	keccak.DefaultKeccakPool.Put(hash)
}

type TransactionHashForkV2 struct {
}

func (th *TransactionHashForkV2) SerializeForRootCalculation(t *Transaction, _ *fastrlp.ArenaPool) []byte {
	return t.MarshalRLPTo(nil)
}

func (th *TransactionHashForkV2) ComputeHash(t *Transaction) {
	var txHash Hash

	hash := keccak.DefaultKeccakPool.Get()
	hash.WriteFn(txHash[:0], t.MarshalRLPTo)
	t.SetHash(txHash)
	keccak.DefaultKeccakPool.Put(hash)
}

func RegisterTxHashFork(txHashWithTypeFork string) error {
	fh := forkmanager.GetInstance()

	if err := fh.RegisterHandler(
		forkmanager.InitialFork, txHashHandler, &TransactionHashForkV1{}); err != nil {
		return err
	}

	if fh.IsForkRegistered(txHashWithTypeFork) {
		if err := fh.RegisterHandler(
			txHashWithTypeFork, txHashHandler, &TransactionHashForkV2{}); err != nil {
			return err
		}
	}

	return nil
}

func GetTransactionHashHandler(blockNumber uint64) TransactionHashFork {
	if h := forkmanager.GetInstance().GetHandler(txHashHandler, blockNumber); h != nil {
		//nolint:forcetypeassert
		return h.(TransactionHashFork)
	}

	// because of tests
	return &TransactionHashForkV1{}
}
