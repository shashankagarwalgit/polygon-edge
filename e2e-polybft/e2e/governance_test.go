package e2e

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/command/sidechain"
	"github.com/0xPolygon/polygon-edge/consensus/polybft"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/contractsapi"
	"github.com/0xPolygon/polygon-edge/contracts"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/e2e-polybft/framework"
	"github.com/0xPolygon/polygon-edge/txrelayer"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

type VoteType uint8
type ProposalState uint8

const (
	Against VoteType = iota
	For
	Abstain
)

const (
	Pending ProposalState = iota
	Active
	Canceled
	Defeated
	Succeeded
	Queued
	Expired
	Executed
)

func TestE2E_Governance_ProposeAndExecuteSimpleProposal(t *testing.T) {
	var (
		oldEpochSize = uint64(10)
		newEpochSize = uint64(20)
		votingPeriod = 2 * oldEpochSize
	)

	cluster := framework.NewTestCluster(t, 5, framework.WithEpochSize(int(oldEpochSize)),
		framework.WithGovernanceVotingPeriod(votingPeriod),
		framework.WithGovernanceVotingDelay(1))
	defer cluster.Stop()

	cluster.WaitForReady(t)

	proposer := cluster.Servers[0] // first validator is governor admin by default in e2e tests

	proposerAcc, err := sidechain.GetAccountFromDir(proposer.DataDir())
	require.NoError(t, err)

	l2Relayer, err := txrelayer.NewTxRelayer(txrelayer.WithClient(proposer.JSONRPC()))
	require.NoError(t, err)

	t.Run("successful change of epoch size", func(t *testing.T) {
		// propose a new epoch size
		setNewEpochSizeFn := &contractsapi.SetNewEpochSizeNetworkParamsFn{
			NewEpochSize: big.NewInt(int64(newEpochSize)),
		}

		proposalInput, err := setNewEpochSizeFn.EncodeAbi()
		require.NoError(t, err)

		proposalDescription := fmt.Sprintf("Change epoch size from %d to %d", oldEpochSize, newEpochSize)

		proposalID := sendProposalTransaction(t, l2Relayer, proposerAcc.Ecdsa,
			proposalInput, proposalDescription)

		fmt.Println(proposalID)

		currentBlockNumber, err := l2Relayer.Client().Eth().BlockNumber()
		require.NoError(t, err)

		// wait couple of blocks until proposal delay finishes
		require.NoError(t, cluster.WaitForBlock(currentBlockNumber+5, time.Minute))

		// vote for the proposal
		for _, s := range cluster.Servers {
			voterAcc, err := sidechain.GetAccountFromDir(s.DataDir())
			require.NoError(t, err)

			sendVoteTransaction(t, proposalID, For, l2Relayer, voterAcc.Ecdsa)
		}

		// wait for voting period to end
		require.NoError(t, cluster.WaitForBlock(votingPeriod+15, 3*time.Minute))

		// check if proposal has quorum (if it was accepted)
		proposalState := getProposalState(t, proposalID, l2Relayer)
		require.Equal(t, Succeeded, proposalState)

		// queue proposal for execution
		sendQueueProposalTransaction(t, l2Relayer, proposerAcc.Ecdsa,
			proposalInput, proposalDescription)

		currentBlockNumber, err = l2Relayer.Client().Eth().BlockNumber()
		require.NoError(t, err)

		// wait for couple of blocks until execution delay passes
		require.NoError(t, cluster.WaitForBlock(currentBlockNumber+3, 30*time.Second))

		// execute proposal
		sendExecuteProposalTransaction(t, l2Relayer, proposerAcc.Ecdsa,
			proposalInput, proposalDescription)

		currentBlockNumber, err = l2Relayer.Client().Eth().BlockNumber()
		require.NoError(t, err)

		// check if epoch size changed on NetworkParams
		networkParamsResponse, err := ABICall(l2Relayer, contractsapi.NetworkParams,
			ethgo.Address(contracts.NetworkParamsContract), ethgo.ZeroAddress, "epochSize")
		require.NoError(t, err)

		epochSizeOnNetworkParams, err := types.ParseUint256orHex(&networkParamsResponse)
		require.NoError(t, err)
		require.Equal(t, newEpochSize, epochSizeOnNetworkParams.Uint64())

		endOfPreviousEpoch := (currentBlockNumber/oldEpochSize + 1) * oldEpochSize
		endOfNewEpoch := endOfPreviousEpoch + newEpochSize

		// wait until the new epoch (with new size finishes)
		require.NoError(t, cluster.WaitForBlock(endOfNewEpoch, 3*time.Minute))

		block, err := l2Relayer.Client().Eth().GetBlockByNumber(
			ethgo.BlockNumber(endOfPreviousEpoch), false)
		require.NoError(t, err)

		extra, err := polybft.GetIbftExtra(block.ExtraData)
		require.NoError(t, err)

		oldEpoch := extra.Checkpoint.EpochNumber

		block, err = l2Relayer.Client().Eth().GetBlockByNumber(
			ethgo.BlockNumber(endOfNewEpoch), false)
		require.NoError(t, err)

		extra, err = polybft.GetIbftExtra(block.ExtraData)
		require.NoError(t, err)

		newEpoch := extra.Checkpoint.EpochNumber

		// check that epochs are sequential
		require.True(t, newEpoch-oldEpoch == 1)
		// check that epoch size actually changed in our consensus
		require.True(t, endOfNewEpoch-endOfPreviousEpoch == newEpochSize)
	})

	t.Run("a proposal does not have enough votes (quorum)", func(t *testing.T) {
		// propose a new sprint size
		setNewSprintSizeFn := &contractsapi.SetNewSprintSizeNetworkParamsFn{
			NewSprintSize: big.NewInt(int64(newEpochSize)),
		}

		proposalInput, err := setNewSprintSizeFn.EncodeAbi()
		require.NoError(t, err)

		proposalDescription := fmt.Sprintf("Change sprint size from %d to %d", oldEpochSize, newEpochSize)

		proposalID := sendProposalTransaction(t, l2Relayer, proposerAcc.Ecdsa,
			proposalInput, proposalDescription)

		fmt.Println(proposalID)

		currentBlockNumber, err := l2Relayer.Client().Eth().BlockNumber()
		require.NoError(t, err)

		// wait couple of blocks until proposal delay finishes
		require.NoError(t, cluster.WaitForBlock(currentBlockNumber+5, time.Minute))

		// only the proposer votes for proposal and rest of them vote against
		sendVoteTransaction(t, proposalID, For, l2Relayer, proposerAcc.Ecdsa)

		for _, s := range cluster.Servers[1:] {
			voterAcc, err := sidechain.GetAccountFromDir(s.DataDir())
			require.NoError(t, err)

			sendVoteTransaction(t, proposalID, Against, l2Relayer, voterAcc.Ecdsa)
		}

		// wait for voting period to end
		require.NoError(t, cluster.WaitForBlock(currentBlockNumber+votingPeriod+5, 3*time.Minute))

		// check if proposal has quorum (if it was accepted), in this case it won't be
		proposalState := getProposalState(t, proposalID, l2Relayer)
		require.Equal(t, Defeated, proposalState)
	})
}

func getProposalState(t *testing.T, proposalID *big.Int, txRelayer txrelayer.TxRelayer) ProposalState {
	t.Helper()

	stateFn := &contractsapi.StateChildGovernorFn{
		ProposalID: proposalID,
	}

	input, err := stateFn.EncodeAbi()
	require.NoError(t, err)

	response, err := txRelayer.Call(ethgo.ZeroAddress, ethgo.Address(contracts.ChildGovernorContract), input)
	require.NoError(t, err)
	require.NotEqual(t, "0x", response)

	converted, err := types.ParseUint8orHex(&response)
	require.NoError(t, err)

	return ProposalState(converted)
}

func sendQueueProposalTransaction(t *testing.T,
	txRelayer txrelayer.TxRelayer, senderKey ethgo.Key,
	input []byte, description string) {
	t.Helper()

	queueFn := contractsapi.QueueChildGovernorFn{
		Targets:         []types.Address{contracts.NetworkParamsContract},
		Calldatas:       [][]byte{input},
		DescriptionHash: crypto.Keccak256Hash([]byte(description)),
		Values:          []*big.Int{big.NewInt(0)},
	}

	input, err := queueFn.EncodeAbi()
	require.NoError(t, err)

	childGovernorAddr := ethgo.Address(contracts.ChildGovernorContract)
	txn := &ethgo.Transaction{
		To:    &childGovernorAddr,
		Input: input,
	}

	receipt, err := txRelayer.SendTransaction(txn, senderKey)
	require.NoError(t, err)
	require.Equal(t, uint64(types.ReceiptSuccess), receipt.Status)
}

func sendExecuteProposalTransaction(t *testing.T,
	txRelayer txrelayer.TxRelayer, senderKey ethgo.Key,
	input []byte, description string) {
	t.Helper()

	executeFn := &contractsapi.ExecuteChildGovernorFn{
		Targets:         []types.Address{contracts.NetworkParamsContract},
		Calldatas:       [][]byte{input},
		DescriptionHash: crypto.Keccak256Hash([]byte(description)),
		Values:          []*big.Int{big.NewInt(0)},
	}

	input, err := executeFn.EncodeAbi()
	require.NoError(t, err)

	childGovernorAddr := ethgo.Address(contracts.ChildGovernorContract)
	txn := &ethgo.Transaction{
		To:    &childGovernorAddr,
		Input: input,
	}

	receipt, err := txRelayer.SendTransaction(txn, senderKey)
	require.NoError(t, err)
	require.Equal(t, uint64(types.ReceiptSuccess), receipt.Status)
}

func sendVoteTransaction(t *testing.T, proposalID *big.Int, vote VoteType,
	txRelayer txrelayer.TxRelayer, senderKey ethgo.Key) {
	t.Helper()

	castVoteFn := &contractsapi.CastVoteChildGovernorFn{
		ProposalID: proposalID,
		Support:    uint8(vote),
	}

	input, err := castVoteFn.EncodeAbi()
	require.NoError(t, err)

	childGovernorAddr := ethgo.Address(contracts.ChildGovernorContract)
	txn := &ethgo.Transaction{
		To:    &childGovernorAddr,
		Input: input,
	}

	receipt, err := txRelayer.SendTransaction(txn, senderKey)
	require.NoError(t, err)
	require.Equal(t, uint64(types.ReceiptSuccess), receipt.Status)
}

func sendProposalTransaction(t *testing.T, txRelayer txrelayer.TxRelayer,
	senderKey ethgo.Key,
	input []byte, description string) *big.Int {
	t.Helper()

	proposeFn := &contractsapi.ProposeChildGovernorFn{
		Targets:     []types.Address{contracts.NetworkParamsContract},
		Calldatas:   [][]byte{input},
		Description: description,
		Values:      []*big.Int{big.NewInt(0)},
	}

	input, err := proposeFn.EncodeAbi()
	require.NoError(t, err)

	childGovernorAddr := ethgo.Address(contracts.ChildGovernorContract)
	txn := &ethgo.Transaction{
		To:    &childGovernorAddr,
		Input: input,
	}

	receipt, err := txRelayer.SendTransaction(txn, senderKey)
	require.NoError(t, err)
	require.Equal(t, uint64(types.ReceiptSuccess), receipt.Status)

	var proposalCreatedEvent contractsapi.ProposalCreatedEvent
	for _, log := range receipt.Logs {
		doesMatch, err := proposalCreatedEvent.ParseLog(log)
		require.NoError(t, err)

		if doesMatch {
			break
		}
	}

	require.NotNil(t, proposalCreatedEvent)

	return proposalCreatedEvent.ProposalID
}