package polybft

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/consensus/polybft/contractsapi"
	"github.com/0xPolygon/polygon-edge/contracts"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state/runtime/precompiled"
	"github.com/0xPolygon/polygon-edge/types"
)

const (
	abiMethodIDLength      = 4
	stTypeBridgeCommitment = "commitment"
	stTypeEndEpoch         = "end-epoch"
)

// PendingCommitment holds merkle trie of bridge transactions accompanied by epoch number
type PendingCommitment struct {
	*contractsapi.Commitment
	MerkleTree *MerkleTree
	Epoch      uint64
}

// NewCommitment creates a new commitment object
func NewCommitment(epoch uint64, stateSyncEvents []*contractsapi.StateSyncedEvent) (*PendingCommitment, error) {
	tree, err := createMerkleTree(stateSyncEvents)
	if err != nil {
		return nil, err
	}

	return &PendingCommitment{
		MerkleTree: tree,
		Epoch:      epoch,
		Commitment: &contractsapi.Commitment{
			StartID: stateSyncEvents[0].ID,
			EndID:   stateSyncEvents[len(stateSyncEvents)-1].ID,
			Root:    tree.Hash(),
		},
	}, nil
}

// Hash calculates hash value for commitment object.
func (cm *PendingCommitment) Hash() (types.Hash, error) {
	data, err := cm.Commitment.EncodeAbi()
	if err != nil {
		return types.Hash{}, err
	}

	return crypto.Keccak256Hash(data), nil
}

var _ contractsapi.StateTransactionInput = &CommitmentMessageSigned{}

// CommitmentMessageSigned encapsulates commitment message with aggregated signatures
type CommitmentMessageSigned struct {
	Message      *contractsapi.Commitment
	AggSignature Signature
	PublicKeys   [][]byte
}

// Hash calculates hash value for commitment object.
func (cm *CommitmentMessageSigned) Hash() (types.Hash, error) {
	data, err := cm.Message.EncodeAbi()
	if err != nil {
		return types.Hash{}, err
	}

	return crypto.Keccak256Hash(data), nil
}

// VerifyStateSyncProof validates given state sync proof
// against merkle trie root hash contained in the CommitmentMessage
func (cm *CommitmentMessageSigned) VerifyStateSyncProof(stateSyncProof *contracts.StateSyncProof) error {
	if stateSyncProof.StateSync == nil {
		return errors.New("no state sync event")
	}

	hash, err := stateSyncProof.StateSync.EncodeAbi()
	if err != nil {
		return err
	}

	return VerifyProof(stateSyncProof.StateSync.ID.Uint64()-cm.Message.StartID.Uint64(),
		hash, stateSyncProof.Proof, cm.Message.Root)
}

// ContainsStateSync checks if commitment contains given state sync event
func (cm *CommitmentMessageSigned) ContainsStateSync(stateSyncID uint64) bool {
	return cm.Message.StartID.Uint64() <= stateSyncID && cm.Message.EndID.Uint64() >= stateSyncID
}

// EncodeAbi contains logic for encoding arbitrary data into ABI format
func (cm *CommitmentMessageSigned) EncodeAbi() ([]byte, error) {
	blsVerificationPart, err := precompiled.BlsVerificationABIType.Encode(
		[2]interface{}{cm.PublicKeys, cm.AggSignature.Bitmap})
	if err != nil {
		return nil, err
	}

	commit := &contractsapi.Commit{
		Commitment: cm.Message,
		Signature:  cm.AggSignature.AggregatedSignature,
		Bitmap:     blsVerificationPart,
	}

	return commit.EncodeAbi()
}

// DecodeAbi contains logic for decoding given ABI data
func (cm *CommitmentMessageSigned) DecodeAbi(txData []byte) error {
	if len(txData) < abiMethodIDLength {
		return fmt.Errorf("invalid commitment data, len = %d", len(txData))
	}

	commit := contractsapi.Commit{}

	err := commit.DecodeAbi(txData)
	if err != nil {
		return err
	}

	decoded, err := precompiled.BlsVerificationABIType.Decode(commit.Bitmap)
	if err != nil {
		return err
	}

	blsMap, isOk := decoded.(map[string]interface{})
	if !isOk {
		return fmt.Errorf("invalid commitment data. Bls verification part not in correct format")
	}

	publicKeys, isOk := blsMap["0"].([][]byte)
	if !isOk {
		return fmt.Errorf("invalid commitment data. Could not find public keys part")
	}

	bitmap, isOk := blsMap["1"].([]byte)
	if !isOk {
		return fmt.Errorf("invalid commitment data. Could not find bitmap part")
	}

	*cm = CommitmentMessageSigned{
		Message: commit.Commitment,
		AggSignature: Signature{
			AggregatedSignature: commit.Signature,
			Bitmap:              bitmap,
		},
		PublicKeys: publicKeys,
	}

	return nil
}

// Type returns type of state transaction input
func (cm *CommitmentMessageSigned) Type() contractsapi.StateTransactionType {
	return stTypeBridgeCommitment
}

func decodeStateTransaction(txData []byte) (contractsapi.StateTransactionInput, error) {
	if len(txData) < abiMethodIDLength {
		return nil, fmt.Errorf("state transactions have input")
	}

	sig := txData[:abiMethodIDLength]

	var obj contractsapi.StateTransactionInput

	if bytes.Equal(sig, contractsapi.StateReceiver.Abi.Methods["commit"].ID()) {
		// bridge commitment
		obj = &CommitmentMessageSigned{}
	} else {
		return nil, fmt.Errorf("unknown state transaction")
	}

	if err := obj.DecodeAbi(txData); err != nil {
		return nil, err
	}

	return obj, nil
}

func getCommitmentMessageSignedTx(txs []*types.Transaction) (*CommitmentMessageSigned, error) {
	for _, tx := range txs {
		// skip non state CommitmentMessageSigned transactions
		if tx.Type != types.StateTx ||
			len(tx.Input) < abiMethodIDLength ||
			!bytes.Equal(tx.Input[:abiMethodIDLength], contractsapi.StateReceiver.Abi.Methods["commit"].ID()) {
			continue
		}

		obj := &CommitmentMessageSigned{}

		if err := obj.DecodeAbi(tx.Input); err != nil {
			return nil, fmt.Errorf("get commitment message signed tx error: %w", err)
		}

		return obj, nil
	}

	return nil, nil
}

func createMerkleTree(stateSyncEvents []*contractsapi.StateSyncedEvent) (*MerkleTree, error) {
	ssh := make([][]byte, len(stateSyncEvents))

	for i, sse := range stateSyncEvents {
		data, err := sse.EncodeAbi()
		if err != nil {
			return nil, err
		}

		ssh[i] = data
	}

	return NewMerkleTree(ssh)
}
