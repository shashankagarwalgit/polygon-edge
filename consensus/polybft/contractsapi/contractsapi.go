// Code generated by scapi/gen. DO NOT EDIT.
package contractsapi

import (
	"math/big"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

type StateSyncCommitment struct {
	StartID *big.Int   `abi:"startId"`
	EndID   *big.Int   `abi:"endId"`
	Root    types.Hash `abi:"root"`
}

var StateSyncCommitmentABIType = abi.MustNewType("tuple(uint256 startId,uint256 endId,bytes32 root)")

func (s *StateSyncCommitment) EncodeAbi() ([]byte, error) {
	return StateSyncCommitmentABIType.Encode(s)
}

func (s *StateSyncCommitment) DecodeAbi(buf []byte) error {
	return decodeStruct(StateSyncCommitmentABIType, buf, &s)
}

type CommitStateReceiverFn struct {
	Commitment *StateSyncCommitment `abi:"commitment"`
	Signature  []byte               `abi:"signature"`
	Bitmap     []byte               `abi:"bitmap"`
}

func (c *CommitStateReceiverFn) Sig() []byte {
	return StateReceiver.Abi.Methods["commit"].ID()
}

func (c *CommitStateReceiverFn) EncodeAbi() ([]byte, error) {
	return StateReceiver.Abi.Methods["commit"].Encode(c)
}

func (c *CommitStateReceiverFn) DecodeAbi(buf []byte) error {
	return decodeMethod(StateReceiver.Abi.Methods["commit"], buf, c)
}

type StateSync struct {
	ID       *big.Int      `abi:"id"`
	Sender   types.Address `abi:"sender"`
	Receiver types.Address `abi:"receiver"`
	Data     []byte        `abi:"data"`
}

var StateSyncABIType = abi.MustNewType("tuple(uint256 id,address sender,address receiver,bytes data)")

func (s *StateSync) EncodeAbi() ([]byte, error) {
	return StateSyncABIType.Encode(s)
}

func (s *StateSync) DecodeAbi(buf []byte) error {
	return decodeStruct(StateSyncABIType, buf, &s)
}

type ExecuteStateReceiverFn struct {
	Proof []types.Hash `abi:"proof"`
	Obj   *StateSync   `abi:"obj"`
}

func (e *ExecuteStateReceiverFn) Sig() []byte {
	return StateReceiver.Abi.Methods["execute"].ID()
}

func (e *ExecuteStateReceiverFn) EncodeAbi() ([]byte, error) {
	return StateReceiver.Abi.Methods["execute"].Encode(e)
}

func (e *ExecuteStateReceiverFn) DecodeAbi(buf []byte) error {
	return decodeMethod(StateReceiver.Abi.Methods["execute"], buf, e)
}

type StateSyncResultEvent struct {
	Counter *big.Int `abi:"counter"`
	Status  bool     `abi:"status"`
	Message []byte   `abi:"message"`
}

func (*StateSyncResultEvent) Sig() ethgo.Hash {
	return StateReceiver.Abi.Events["StateSyncResult"].ID()
}

func (*StateSyncResultEvent) Encode(inputs interface{}) ([]byte, error) {
	return StateReceiver.Abi.Events["StateSyncResult"].Inputs.Encode(inputs)
}

func (s *StateSyncResultEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !StateReceiver.Abi.Events["StateSyncResult"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(StateReceiver.Abi.Events["StateSyncResult"], log, s)
}

type NewCommitmentEvent struct {
	StartID *big.Int   `abi:"startId"`
	EndID   *big.Int   `abi:"endId"`
	Root    types.Hash `abi:"root"`
}

func (*NewCommitmentEvent) Sig() ethgo.Hash {
	return StateReceiver.Abi.Events["NewCommitment"].ID()
}

func (*NewCommitmentEvent) Encode(inputs interface{}) ([]byte, error) {
	return StateReceiver.Abi.Events["NewCommitment"].Inputs.Encode(inputs)
}

func (n *NewCommitmentEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !StateReceiver.Abi.Events["NewCommitment"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(StateReceiver.Abi.Events["NewCommitment"], log, n)
}

type Epoch struct {
	StartBlock *big.Int   `abi:"startBlock"`
	EndBlock   *big.Int   `abi:"endBlock"`
	EpochRoot  types.Hash `abi:"epochRoot"`
}

var EpochABIType = abi.MustNewType("tuple(uint256 startBlock,uint256 endBlock,bytes32 epochRoot)")

func (e *Epoch) EncodeAbi() ([]byte, error) {
	return EpochABIType.Encode(e)
}

func (e *Epoch) DecodeAbi(buf []byte) error {
	return decodeStruct(EpochABIType, buf, &e)
}

type UptimeData struct {
	Validator    types.Address `abi:"validator"`
	SignedBlocks *big.Int      `abi:"signedBlocks"`
}

var UptimeDataABIType = abi.MustNewType("tuple(address validator,uint256 signedBlocks)")

func (u *UptimeData) EncodeAbi() ([]byte, error) {
	return UptimeDataABIType.Encode(u)
}

func (u *UptimeData) DecodeAbi(buf []byte) error {
	return decodeStruct(UptimeDataABIType, buf, &u)
}

type Uptime struct {
	EpochID     *big.Int      `abi:"epochId"`
	UptimeData  []*UptimeData `abi:"uptimeData"`
	TotalBlocks *big.Int      `abi:"totalBlocks"`
}

var UptimeABIType = abi.MustNewType("tuple(uint256 epochId,tuple(address validator,uint256 signedBlocks)[] uptimeData,uint256 totalBlocks)")

func (u *Uptime) EncodeAbi() ([]byte, error) {
	return UptimeABIType.Encode(u)
}

func (u *Uptime) DecodeAbi(buf []byte) error {
	return decodeStruct(UptimeABIType, buf, &u)
}

type CommitEpochChildValidatorSetFn struct {
	ID     *big.Int `abi:"id"`
	Epoch  *Epoch   `abi:"epoch"`
	Uptime *Uptime  `abi:"uptime"`
}

func (c *CommitEpochChildValidatorSetFn) Sig() []byte {
	return ChildValidatorSet.Abi.Methods["commitEpoch"].ID()
}

func (c *CommitEpochChildValidatorSetFn) EncodeAbi() ([]byte, error) {
	return ChildValidatorSet.Abi.Methods["commitEpoch"].Encode(c)
}

func (c *CommitEpochChildValidatorSetFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildValidatorSet.Abi.Methods["commitEpoch"], buf, c)
}

type InitStruct struct {
	EpochReward   *big.Int `abi:"epochReward"`
	MinStake      *big.Int `abi:"minStake"`
	MinDelegation *big.Int `abi:"minDelegation"`
	EpochSize     *big.Int `abi:"epochSize"`
}

var InitStructABIType = abi.MustNewType("tuple(uint256 epochReward,uint256 minStake,uint256 minDelegation,uint256 epochSize)")

func (i *InitStruct) EncodeAbi() ([]byte, error) {
	return InitStructABIType.Encode(i)
}

func (i *InitStruct) DecodeAbi(buf []byte) error {
	return decodeStruct(InitStructABIType, buf, &i)
}

type ValidatorInit struct {
	Addr      types.Address `abi:"addr"`
	Pubkey    [4]*big.Int   `abi:"pubkey"`
	Signature [2]*big.Int   `abi:"signature"`
	Stake     *big.Int      `abi:"stake"`
}

var ValidatorInitABIType = abi.MustNewType("tuple(address addr,uint256[4] pubkey,uint256[2] signature,uint256 stake)")

func (v *ValidatorInit) EncodeAbi() ([]byte, error) {
	return ValidatorInitABIType.Encode(v)
}

func (v *ValidatorInit) DecodeAbi(buf []byte) error {
	return decodeStruct(ValidatorInitABIType, buf, &v)
}

type InitializeChildValidatorSetFn struct {
	Init       *InitStruct      `abi:"init"`
	Validators []*ValidatorInit `abi:"validators"`
	NewBls     types.Address    `abi:"newBls"`
	Governance types.Address    `abi:"governance"`
}

func (i *InitializeChildValidatorSetFn) Sig() []byte {
	return ChildValidatorSet.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildValidatorSetFn) EncodeAbi() ([]byte, error) {
	return ChildValidatorSet.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildValidatorSetFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildValidatorSet.Abi.Methods["initialize"], buf, i)
}

type AddToWhitelistChildValidatorSetFn struct {
	WhitelistAddreses []ethgo.Address `abi:"whitelistAddreses"`
}

func (a *AddToWhitelistChildValidatorSetFn) Sig() []byte {
	return ChildValidatorSet.Abi.Methods["addToWhitelist"].ID()
}

func (a *AddToWhitelistChildValidatorSetFn) EncodeAbi() ([]byte, error) {
	return ChildValidatorSet.Abi.Methods["addToWhitelist"].Encode(a)
}

func (a *AddToWhitelistChildValidatorSetFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildValidatorSet.Abi.Methods["addToWhitelist"], buf, a)
}

type RegisterChildValidatorSetFn struct {
	Signature [2]*big.Int `abi:"signature"`
	Pubkey    [4]*big.Int `abi:"pubkey"`
}

func (r *RegisterChildValidatorSetFn) Sig() []byte {
	return ChildValidatorSet.Abi.Methods["register"].ID()
}

func (r *RegisterChildValidatorSetFn) EncodeAbi() ([]byte, error) {
	return ChildValidatorSet.Abi.Methods["register"].Encode(r)
}

func (r *RegisterChildValidatorSetFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildValidatorSet.Abi.Methods["register"], buf, r)
}

type NewValidatorEvent struct {
	Validator types.Address `abi:"validator"`
	BlsKey    [4]*big.Int   `abi:"blsKey"`
}

func (*NewValidatorEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["NewValidator"].ID()
}

func (*NewValidatorEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["NewValidator"].Inputs.Encode(inputs)
}

func (n *NewValidatorEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["NewValidator"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["NewValidator"], log, n)
}

type StakedEvent struct {
	Validator types.Address `abi:"validator"`
	Amount    *big.Int      `abi:"amount"`
}

func (*StakedEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["Staked"].ID()
}

func (*StakedEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["Staked"].Inputs.Encode(inputs)
}

func (s *StakedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["Staked"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["Staked"], log, s)
}

type DelegatedEvent struct {
	Delegator types.Address `abi:"delegator"`
	Validator types.Address `abi:"validator"`
	Amount    *big.Int      `abi:"amount"`
}

func (*DelegatedEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["Delegated"].ID()
}

func (*DelegatedEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["Delegated"].Inputs.Encode(inputs)
}

func (d *DelegatedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["Delegated"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["Delegated"], log, d)
}

type UnstakedEvent struct {
	Validator types.Address `abi:"validator"`
	Amount    *big.Int      `abi:"amount"`
}

func (*UnstakedEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["Unstaked"].ID()
}

func (*UnstakedEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["Unstaked"].Inputs.Encode(inputs)
}

func (u *UnstakedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["Unstaked"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["Unstaked"], log, u)
}

type UndelegatedEvent struct {
	Delegator types.Address `abi:"delegator"`
	Validator types.Address `abi:"validator"`
	Amount    *big.Int      `abi:"amount"`
}

func (*UndelegatedEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["Undelegated"].ID()
}

func (*UndelegatedEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["Undelegated"].Inputs.Encode(inputs)
}

func (u *UndelegatedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["Undelegated"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["Undelegated"], log, u)
}

type AddedToWhitelistEvent struct {
	Validator types.Address `abi:"validator"`
}

func (*AddedToWhitelistEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["AddedToWhitelist"].ID()
}

func (*AddedToWhitelistEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["AddedToWhitelist"].Inputs.Encode(inputs)
}

func (a *AddedToWhitelistEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["AddedToWhitelist"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["AddedToWhitelist"], log, a)
}

type WithdrawalEvent struct {
	Account types.Address `abi:"account"`
	To      types.Address `abi:"to"`
	Amount  *big.Int      `abi:"amount"`
}

func (*WithdrawalEvent) Sig() ethgo.Hash {
	return ChildValidatorSet.Abi.Events["Withdrawal"].ID()
}

func (*WithdrawalEvent) Encode(inputs interface{}) ([]byte, error) {
	return ChildValidatorSet.Abi.Events["Withdrawal"].Inputs.Encode(inputs)
}

func (w *WithdrawalEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !ChildValidatorSet.Abi.Events["Withdrawal"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(ChildValidatorSet.Abi.Events["Withdrawal"], log, w)
}

type SyncStateStateSenderFn struct {
	Receiver types.Address `abi:"receiver"`
	Data     []byte        `abi:"data"`
}

func (s *SyncStateStateSenderFn) Sig() []byte {
	return StateSender.Abi.Methods["syncState"].ID()
}

func (s *SyncStateStateSenderFn) EncodeAbi() ([]byte, error) {
	return StateSender.Abi.Methods["syncState"].Encode(s)
}

func (s *SyncStateStateSenderFn) DecodeAbi(buf []byte) error {
	return decodeMethod(StateSender.Abi.Methods["syncState"], buf, s)
}

type StateSyncedEvent struct {
	ID       *big.Int      `abi:"id"`
	Sender   types.Address `abi:"sender"`
	Receiver types.Address `abi:"receiver"`
	Data     []byte        `abi:"data"`
}

func (*StateSyncedEvent) Sig() ethgo.Hash {
	return StateSender.Abi.Events["StateSynced"].ID()
}

func (*StateSyncedEvent) Encode(inputs interface{}) ([]byte, error) {
	return StateSender.Abi.Events["StateSynced"].Inputs.Encode(inputs)
}

func (s *StateSyncedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !StateSender.Abi.Events["StateSynced"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(StateSender.Abi.Events["StateSynced"], log, s)
}

type L2StateSyncedEvent struct {
	ID       *big.Int      `abi:"id"`
	Sender   types.Address `abi:"sender"`
	Receiver types.Address `abi:"receiver"`
	Data     []byte        `abi:"data"`
}

func (*L2StateSyncedEvent) Sig() ethgo.Hash {
	return L2StateSender.Abi.Events["L2StateSynced"].ID()
}

func (*L2StateSyncedEvent) Encode(inputs interface{}) ([]byte, error) {
	return L2StateSender.Abi.Events["L2StateSynced"].Inputs.Encode(inputs)
}

func (l *L2StateSyncedEvent) ParseLog(log *ethgo.Log) (bool, error) {
	if !L2StateSender.Abi.Events["L2StateSynced"].Match(log) {
		return false, nil
	}

	return true, decodeEvent(L2StateSender.Abi.Events["L2StateSynced"], log, l)
}

type CheckpointMetadata struct {
	BlockHash               types.Hash `abi:"blockHash"`
	BlockRound              *big.Int   `abi:"blockRound"`
	CurrentValidatorSetHash types.Hash `abi:"currentValidatorSetHash"`
}

var CheckpointMetadataABIType = abi.MustNewType("tuple(bytes32 blockHash,uint256 blockRound,bytes32 currentValidatorSetHash)")

func (c *CheckpointMetadata) EncodeAbi() ([]byte, error) {
	return CheckpointMetadataABIType.Encode(c)
}

func (c *CheckpointMetadata) DecodeAbi(buf []byte) error {
	return decodeStruct(CheckpointMetadataABIType, buf, &c)
}

type Checkpoint struct {
	Epoch       *big.Int   `abi:"epoch"`
	BlockNumber *big.Int   `abi:"blockNumber"`
	EventRoot   types.Hash `abi:"eventRoot"`
}

var CheckpointABIType = abi.MustNewType("tuple(uint256 epoch,uint256 blockNumber,bytes32 eventRoot)")

func (c *Checkpoint) EncodeAbi() ([]byte, error) {
	return CheckpointABIType.Encode(c)
}

func (c *Checkpoint) DecodeAbi(buf []byte) error {
	return decodeStruct(CheckpointABIType, buf, &c)
}

type Validator struct {
	Address     types.Address `abi:"_address"`
	BlsKey      [4]*big.Int   `abi:"blsKey"`
	VotingPower *big.Int      `abi:"votingPower"`
}

var ValidatorABIType = abi.MustNewType("tuple(address _address,uint256[4] blsKey,uint256 votingPower)")

func (v *Validator) EncodeAbi() ([]byte, error) {
	return ValidatorABIType.Encode(v)
}

func (v *Validator) DecodeAbi(buf []byte) error {
	return decodeStruct(ValidatorABIType, buf, &v)
}

type SubmitCheckpointManagerFn struct {
	CheckpointMetadata *CheckpointMetadata `abi:"checkpointMetadata"`
	Checkpoint         *Checkpoint         `abi:"checkpoint"`
	Signature          [2]*big.Int         `abi:"signature"`
	NewValidatorSet    []*Validator        `abi:"newValidatorSet"`
	Bitmap             []byte              `abi:"bitmap"`
}

func (s *SubmitCheckpointManagerFn) Sig() []byte {
	return CheckpointManager.Abi.Methods["submit"].ID()
}

func (s *SubmitCheckpointManagerFn) EncodeAbi() ([]byte, error) {
	return CheckpointManager.Abi.Methods["submit"].Encode(s)
}

func (s *SubmitCheckpointManagerFn) DecodeAbi(buf []byte) error {
	return decodeMethod(CheckpointManager.Abi.Methods["submit"], buf, s)
}

type InitializeCheckpointManagerFn struct {
	NewBls          types.Address `abi:"newBls"`
	NewBn256G2      types.Address `abi:"newBn256G2"`
	ChainID_        *big.Int      `abi:"chainId_"`
	NewValidatorSet []*Validator  `abi:"newValidatorSet"`
}

func (i *InitializeCheckpointManagerFn) Sig() []byte {
	return CheckpointManager.Abi.Methods["initialize"].ID()
}

func (i *InitializeCheckpointManagerFn) EncodeAbi() ([]byte, error) {
	return CheckpointManager.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeCheckpointManagerFn) DecodeAbi(buf []byte) error {
	return decodeMethod(CheckpointManager.Abi.Methods["initialize"], buf, i)
}

type GetCheckpointBlockCheckpointManagerFn struct {
	BlockNumber *big.Int `abi:"blockNumber"`
}

func (g *GetCheckpointBlockCheckpointManagerFn) Sig() []byte {
	return CheckpointManager.Abi.Methods["getCheckpointBlock"].ID()
}

func (g *GetCheckpointBlockCheckpointManagerFn) EncodeAbi() ([]byte, error) {
	return CheckpointManager.Abi.Methods["getCheckpointBlock"].Encode(g)
}

func (g *GetCheckpointBlockCheckpointManagerFn) DecodeAbi(buf []byte) error {
	return decodeMethod(CheckpointManager.Abi.Methods["getCheckpointBlock"], buf, g)
}

type InitializeExitHelperFn struct {
	NewCheckpointManager types.Address `abi:"newCheckpointManager"`
}

func (i *InitializeExitHelperFn) Sig() []byte {
	return ExitHelper.Abi.Methods["initialize"].ID()
}

func (i *InitializeExitHelperFn) EncodeAbi() ([]byte, error) {
	return ExitHelper.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeExitHelperFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ExitHelper.Abi.Methods["initialize"], buf, i)
}

type ExitExitHelperFn struct {
	BlockNumber  *big.Int     `abi:"blockNumber"`
	LeafIndex    *big.Int     `abi:"leafIndex"`
	UnhashedLeaf []byte       `abi:"unhashedLeaf"`
	Proof        []types.Hash `abi:"proof"`
}

func (e *ExitExitHelperFn) Sig() []byte {
	return ExitHelper.Abi.Methods["exit"].ID()
}

func (e *ExitExitHelperFn) EncodeAbi() ([]byte, error) {
	return ExitHelper.Abi.Methods["exit"].Encode(e)
}

func (e *ExitExitHelperFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ExitHelper.Abi.Methods["exit"], buf, e)
}

type InitializeChildERC20PredicateFn struct {
	NewL2StateSender          types.Address `abi:"newL2StateSender"`
	NewStateReceiver          types.Address `abi:"newStateReceiver"`
	NewRootERC20Predicate     types.Address `abi:"newRootERC20Predicate"`
	NewChildTokenTemplate     types.Address `abi:"newChildTokenTemplate"`
	NewNativeTokenRootAddress types.Address `abi:"newNativeTokenRootAddress"`
}

func (i *InitializeChildERC20PredicateFn) Sig() []byte {
	return ChildERC20Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC20PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC20Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC20PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC20Predicate.Abi.Methods["initialize"], buf, i)
}

type WithdrawToChildERC20PredicateFn struct {
	ChildToken types.Address `abi:"childToken"`
	Receiver   types.Address `abi:"receiver"`
	Amount     *big.Int      `abi:"amount"`
}

func (w *WithdrawToChildERC20PredicateFn) Sig() []byte {
	return ChildERC20Predicate.Abi.Methods["withdrawTo"].ID()
}

func (w *WithdrawToChildERC20PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC20Predicate.Abi.Methods["withdrawTo"].Encode(w)
}

func (w *WithdrawToChildERC20PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC20Predicate.Abi.Methods["withdrawTo"], buf, w)
}

type InitializeChildERC20PredicateAccessListFn struct {
	NewL2StateSender          types.Address `abi:"newL2StateSender"`
	NewStateReceiver          types.Address `abi:"newStateReceiver"`
	NewRootERC20Predicate     types.Address `abi:"newRootERC20Predicate"`
	NewChildTokenTemplate     types.Address `abi:"newChildTokenTemplate"`
	NewNativeTokenRootAddress types.Address `abi:"newNativeTokenRootAddress"`
	UseAllowList              bool          `abi:"useAllowList"`
	UseBlockList              bool          `abi:"useBlockList"`
	NewOwner                  types.Address `abi:"newOwner"`
}

func (i *InitializeChildERC20PredicateAccessListFn) Sig() []byte {
	return ChildERC20PredicateAccessList.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC20PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC20PredicateAccessList.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC20PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC20PredicateAccessList.Abi.Methods["initialize"], buf, i)
}

type WithdrawToChildERC20PredicateAccessListFn struct {
	ChildToken types.Address `abi:"childToken"`
	Receiver   types.Address `abi:"receiver"`
	Amount     *big.Int      `abi:"amount"`
}

func (w *WithdrawToChildERC20PredicateAccessListFn) Sig() []byte {
	return ChildERC20PredicateAccessList.Abi.Methods["withdrawTo"].ID()
}

func (w *WithdrawToChildERC20PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC20PredicateAccessList.Abi.Methods["withdrawTo"].Encode(w)
}

func (w *WithdrawToChildERC20PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC20PredicateAccessList.Abi.Methods["withdrawTo"], buf, w)
}

type InitializeNativeERC20Fn struct {
	Predicate_ types.Address `abi:"predicate_"`
	RootToken_ types.Address `abi:"rootToken_"`
	Name_      string        `abi:"name_"`
	Symbol_    string        `abi:"symbol_"`
	Decimals_  uint8         `abi:"decimals_"`
}

func (i *InitializeNativeERC20Fn) Sig() []byte {
	return NativeERC20.Abi.Methods["initialize"].ID()
}

func (i *InitializeNativeERC20Fn) EncodeAbi() ([]byte, error) {
	return NativeERC20.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeNativeERC20Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(NativeERC20.Abi.Methods["initialize"], buf, i)
}

type InitializeNativeERC20MintableFn struct {
	Predicate_ types.Address `abi:"predicate_"`
	Owner_     types.Address `abi:"owner_"`
	RootToken_ types.Address `abi:"rootToken_"`
	Name_      string        `abi:"name_"`
	Symbol_    string        `abi:"symbol_"`
	Decimals_  uint8         `abi:"decimals_"`
}

func (i *InitializeNativeERC20MintableFn) Sig() []byte {
	return NativeERC20Mintable.Abi.Methods["initialize"].ID()
}

func (i *InitializeNativeERC20MintableFn) EncodeAbi() ([]byte, error) {
	return NativeERC20Mintable.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeNativeERC20MintableFn) DecodeAbi(buf []byte) error {
	return decodeMethod(NativeERC20Mintable.Abi.Methods["initialize"], buf, i)
}

type InitializeRootERC20PredicateFn struct {
	NewStateSender         types.Address `abi:"newStateSender"`
	NewExitHelper          types.Address `abi:"newExitHelper"`
	NewChildERC20Predicate types.Address `abi:"newChildERC20Predicate"`
	NewChildTokenTemplate  types.Address `abi:"newChildTokenTemplate"`
	NativeTokenRootAddress types.Address `abi:"nativeTokenRootAddress"`
}

func (i *InitializeRootERC20PredicateFn) Sig() []byte {
	return RootERC20Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeRootERC20PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC20Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeRootERC20PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC20Predicate.Abi.Methods["initialize"], buf, i)
}

type DepositToRootERC20PredicateFn struct {
	RootToken types.Address `abi:"rootToken"`
	Receiver  types.Address `abi:"receiver"`
	Amount    *big.Int      `abi:"amount"`
}

func (d *DepositToRootERC20PredicateFn) Sig() []byte {
	return RootERC20Predicate.Abi.Methods["depositTo"].ID()
}

func (d *DepositToRootERC20PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC20Predicate.Abi.Methods["depositTo"].Encode(d)
}

func (d *DepositToRootERC20PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC20Predicate.Abi.Methods["depositTo"], buf, d)
}

type BalanceOfRootERC20Fn struct {
	Account types.Address `abi:"account"`
}

func (b *BalanceOfRootERC20Fn) Sig() []byte {
	return RootERC20.Abi.Methods["balanceOf"].ID()
}

func (b *BalanceOfRootERC20Fn) EncodeAbi() ([]byte, error) {
	return RootERC20.Abi.Methods["balanceOf"].Encode(b)
}

func (b *BalanceOfRootERC20Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC20.Abi.Methods["balanceOf"], buf, b)
}

type ApproveRootERC20Fn struct {
	Spender types.Address `abi:"spender"`
	Amount  *big.Int      `abi:"amount"`
}

func (a *ApproveRootERC20Fn) Sig() []byte {
	return RootERC20.Abi.Methods["approve"].ID()
}

func (a *ApproveRootERC20Fn) EncodeAbi() ([]byte, error) {
	return RootERC20.Abi.Methods["approve"].Encode(a)
}

func (a *ApproveRootERC20Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC20.Abi.Methods["approve"], buf, a)
}

type MintRootERC20Fn struct {
	To     types.Address `abi:"to"`
	Amount *big.Int      `abi:"amount"`
}

func (m *MintRootERC20Fn) Sig() []byte {
	return RootERC20.Abi.Methods["mint"].ID()
}

func (m *MintRootERC20Fn) EncodeAbi() ([]byte, error) {
	return RootERC20.Abi.Methods["mint"].Encode(m)
}

func (m *MintRootERC20Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC20.Abi.Methods["mint"], buf, m)
}

type InitializeRootERC1155PredicateFn struct {
	NewStateSender           types.Address `abi:"newStateSender"`
	NewExitHelper            types.Address `abi:"newExitHelper"`
	NewChildERC1155Predicate types.Address `abi:"newChildERC1155Predicate"`
	NewChildTokenTemplate    types.Address `abi:"newChildTokenTemplate"`
}

func (i *InitializeRootERC1155PredicateFn) Sig() []byte {
	return RootERC1155Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeRootERC1155PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC1155Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeRootERC1155PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC1155Predicate.Abi.Methods["initialize"], buf, i)
}

type DepositBatchRootERC1155PredicateFn struct {
	RootToken types.Address   `abi:"rootToken"`
	Receivers []ethgo.Address `abi:"receivers"`
	TokenIDs  []*big.Int      `abi:"tokenIds"`
	Amounts   []*big.Int      `abi:"amounts"`
}

func (d *DepositBatchRootERC1155PredicateFn) Sig() []byte {
	return RootERC1155Predicate.Abi.Methods["depositBatch"].ID()
}

func (d *DepositBatchRootERC1155PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC1155Predicate.Abi.Methods["depositBatch"].Encode(d)
}

func (d *DepositBatchRootERC1155PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC1155Predicate.Abi.Methods["depositBatch"], buf, d)
}

type SetApprovalForAllRootERC1155Fn struct {
	Operator types.Address `abi:"operator"`
	Approved bool          `abi:"approved"`
}

func (s *SetApprovalForAllRootERC1155Fn) Sig() []byte {
	return RootERC1155.Abi.Methods["setApprovalForAll"].ID()
}

func (s *SetApprovalForAllRootERC1155Fn) EncodeAbi() ([]byte, error) {
	return RootERC1155.Abi.Methods["setApprovalForAll"].Encode(s)
}

func (s *SetApprovalForAllRootERC1155Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC1155.Abi.Methods["setApprovalForAll"], buf, s)
}

type MintBatchRootERC1155Fn struct {
	To      types.Address `abi:"to"`
	IDs     []*big.Int    `abi:"ids"`
	Amounts []*big.Int    `abi:"amounts"`
	Data    []byte        `abi:"data"`
}

func (m *MintBatchRootERC1155Fn) Sig() []byte {
	return RootERC1155.Abi.Methods["mintBatch"].ID()
}

func (m *MintBatchRootERC1155Fn) EncodeAbi() ([]byte, error) {
	return RootERC1155.Abi.Methods["mintBatch"].Encode(m)
}

func (m *MintBatchRootERC1155Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC1155.Abi.Methods["mintBatch"], buf, m)
}

type BalanceOfRootERC1155Fn struct {
	Account types.Address `abi:"account"`
	ID      *big.Int      `abi:"id"`
}

func (b *BalanceOfRootERC1155Fn) Sig() []byte {
	return RootERC1155.Abi.Methods["balanceOf"].ID()
}

func (b *BalanceOfRootERC1155Fn) EncodeAbi() ([]byte, error) {
	return RootERC1155.Abi.Methods["balanceOf"].Encode(b)
}

func (b *BalanceOfRootERC1155Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC1155.Abi.Methods["balanceOf"], buf, b)
}

type InitializeChildERC1155PredicateFn struct {
	NewL2StateSender        types.Address `abi:"newL2StateSender"`
	NewStateReceiver        types.Address `abi:"newStateReceiver"`
	NewRootERC1155Predicate types.Address `abi:"newRootERC1155Predicate"`
	NewChildTokenTemplate   types.Address `abi:"newChildTokenTemplate"`
}

func (i *InitializeChildERC1155PredicateFn) Sig() []byte {
	return ChildERC1155Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC1155PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC1155Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC1155PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155Predicate.Abi.Methods["initialize"], buf, i)
}

type WithdrawBatchChildERC1155PredicateFn struct {
	ChildToken types.Address   `abi:"childToken"`
	Receivers  []ethgo.Address `abi:"receivers"`
	TokenIDs   []*big.Int      `abi:"tokenIds"`
	Amounts    []*big.Int      `abi:"amounts"`
}

func (w *WithdrawBatchChildERC1155PredicateFn) Sig() []byte {
	return ChildERC1155Predicate.Abi.Methods["withdrawBatch"].ID()
}

func (w *WithdrawBatchChildERC1155PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC1155Predicate.Abi.Methods["withdrawBatch"].Encode(w)
}

func (w *WithdrawBatchChildERC1155PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155Predicate.Abi.Methods["withdrawBatch"], buf, w)
}

type InitializeChildERC1155PredicateAccessListFn struct {
	NewL2StateSender        types.Address `abi:"newL2StateSender"`
	NewStateReceiver        types.Address `abi:"newStateReceiver"`
	NewRootERC1155Predicate types.Address `abi:"newRootERC1155Predicate"`
	NewChildTokenTemplate   types.Address `abi:"newChildTokenTemplate"`
	UseAllowList            bool          `abi:"useAllowList"`
	UseBlockList            bool          `abi:"useBlockList"`
	NewOwner                types.Address `abi:"newOwner"`
}

func (i *InitializeChildERC1155PredicateAccessListFn) Sig() []byte {
	return ChildERC1155PredicateAccessList.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC1155PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC1155PredicateAccessList.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC1155PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155PredicateAccessList.Abi.Methods["initialize"], buf, i)
}

type WithdrawBatchChildERC1155PredicateAccessListFn struct {
	ChildToken types.Address   `abi:"childToken"`
	Receivers  []ethgo.Address `abi:"receivers"`
	TokenIDs   []*big.Int      `abi:"tokenIds"`
	Amounts    []*big.Int      `abi:"amounts"`
}

func (w *WithdrawBatchChildERC1155PredicateAccessListFn) Sig() []byte {
	return ChildERC1155PredicateAccessList.Abi.Methods["withdrawBatch"].ID()
}

func (w *WithdrawBatchChildERC1155PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC1155PredicateAccessList.Abi.Methods["withdrawBatch"].Encode(w)
}

func (w *WithdrawBatchChildERC1155PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155PredicateAccessList.Abi.Methods["withdrawBatch"], buf, w)
}

type InitializeChildERC1155Fn struct {
	RootToken_ types.Address `abi:"rootToken_"`
	Uri_       string        `abi:"uri_"`
}

func (i *InitializeChildERC1155Fn) Sig() []byte {
	return ChildERC1155.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC1155Fn) EncodeAbi() ([]byte, error) {
	return ChildERC1155.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC1155Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155.Abi.Methods["initialize"], buf, i)
}

type BalanceOfChildERC1155Fn struct {
	Account types.Address `abi:"account"`
	ID      *big.Int      `abi:"id"`
}

func (b *BalanceOfChildERC1155Fn) Sig() []byte {
	return ChildERC1155.Abi.Methods["balanceOf"].ID()
}

func (b *BalanceOfChildERC1155Fn) EncodeAbi() ([]byte, error) {
	return ChildERC1155.Abi.Methods["balanceOf"].Encode(b)
}

func (b *BalanceOfChildERC1155Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC1155.Abi.Methods["balanceOf"], buf, b)
}

type InitializeRootERC721PredicateFn struct {
	NewStateSender          types.Address `abi:"newStateSender"`
	NewExitHelper           types.Address `abi:"newExitHelper"`
	NewChildERC721Predicate types.Address `abi:"newChildERC721Predicate"`
	NewChildTokenTemplate   types.Address `abi:"newChildTokenTemplate"`
}

func (i *InitializeRootERC721PredicateFn) Sig() []byte {
	return RootERC721Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeRootERC721PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC721Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeRootERC721PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC721Predicate.Abi.Methods["initialize"], buf, i)
}

type DepositBatchRootERC721PredicateFn struct {
	RootToken types.Address   `abi:"rootToken"`
	Receivers []ethgo.Address `abi:"receivers"`
	TokenIDs  []*big.Int      `abi:"tokenIds"`
}

func (d *DepositBatchRootERC721PredicateFn) Sig() []byte {
	return RootERC721Predicate.Abi.Methods["depositBatch"].ID()
}

func (d *DepositBatchRootERC721PredicateFn) EncodeAbi() ([]byte, error) {
	return RootERC721Predicate.Abi.Methods["depositBatch"].Encode(d)
}

func (d *DepositBatchRootERC721PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC721Predicate.Abi.Methods["depositBatch"], buf, d)
}

type SetApprovalForAllRootERC721Fn struct {
	Operator types.Address `abi:"operator"`
	Approved bool          `abi:"approved"`
}

func (s *SetApprovalForAllRootERC721Fn) Sig() []byte {
	return RootERC721.Abi.Methods["setApprovalForAll"].ID()
}

func (s *SetApprovalForAllRootERC721Fn) EncodeAbi() ([]byte, error) {
	return RootERC721.Abi.Methods["setApprovalForAll"].Encode(s)
}

func (s *SetApprovalForAllRootERC721Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC721.Abi.Methods["setApprovalForAll"], buf, s)
}

type MintRootERC721Fn struct {
	To types.Address `abi:"to"`
}

func (m *MintRootERC721Fn) Sig() []byte {
	return RootERC721.Abi.Methods["mint"].ID()
}

func (m *MintRootERC721Fn) EncodeAbi() ([]byte, error) {
	return RootERC721.Abi.Methods["mint"].Encode(m)
}

func (m *MintRootERC721Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(RootERC721.Abi.Methods["mint"], buf, m)
}

type InitializeChildERC721PredicateFn struct {
	NewL2StateSender       types.Address `abi:"newL2StateSender"`
	NewStateReceiver       types.Address `abi:"newStateReceiver"`
	NewRootERC721Predicate types.Address `abi:"newRootERC721Predicate"`
	NewChildTokenTemplate  types.Address `abi:"newChildTokenTemplate"`
}

func (i *InitializeChildERC721PredicateFn) Sig() []byte {
	return ChildERC721Predicate.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC721PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC721Predicate.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC721PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721Predicate.Abi.Methods["initialize"], buf, i)
}

type WithdrawBatchChildERC721PredicateFn struct {
	ChildToken types.Address   `abi:"childToken"`
	Receivers  []ethgo.Address `abi:"receivers"`
	TokenIDs   []*big.Int      `abi:"tokenIds"`
}

func (w *WithdrawBatchChildERC721PredicateFn) Sig() []byte {
	return ChildERC721Predicate.Abi.Methods["withdrawBatch"].ID()
}

func (w *WithdrawBatchChildERC721PredicateFn) EncodeAbi() ([]byte, error) {
	return ChildERC721Predicate.Abi.Methods["withdrawBatch"].Encode(w)
}

func (w *WithdrawBatchChildERC721PredicateFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721Predicate.Abi.Methods["withdrawBatch"], buf, w)
}

type InitializeChildERC721PredicateAccessListFn struct {
	NewL2StateSender       types.Address `abi:"newL2StateSender"`
	NewStateReceiver       types.Address `abi:"newStateReceiver"`
	NewRootERC721Predicate types.Address `abi:"newRootERC721Predicate"`
	NewChildTokenTemplate  types.Address `abi:"newChildTokenTemplate"`
	UseAllowList           bool          `abi:"useAllowList"`
	UseBlockList           bool          `abi:"useBlockList"`
	NewOwner               types.Address `abi:"newOwner"`
}

func (i *InitializeChildERC721PredicateAccessListFn) Sig() []byte {
	return ChildERC721PredicateAccessList.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC721PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC721PredicateAccessList.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC721PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721PredicateAccessList.Abi.Methods["initialize"], buf, i)
}

type WithdrawBatchChildERC721PredicateAccessListFn struct {
	ChildToken types.Address   `abi:"childToken"`
	Receivers  []ethgo.Address `abi:"receivers"`
	TokenIDs   []*big.Int      `abi:"tokenIds"`
}

func (w *WithdrawBatchChildERC721PredicateAccessListFn) Sig() []byte {
	return ChildERC721PredicateAccessList.Abi.Methods["withdrawBatch"].ID()
}

func (w *WithdrawBatchChildERC721PredicateAccessListFn) EncodeAbi() ([]byte, error) {
	return ChildERC721PredicateAccessList.Abi.Methods["withdrawBatch"].Encode(w)
}

func (w *WithdrawBatchChildERC721PredicateAccessListFn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721PredicateAccessList.Abi.Methods["withdrawBatch"], buf, w)
}

type InitializeChildERC721Fn struct {
	RootToken_ types.Address `abi:"rootToken_"`
	Name_      string        `abi:"name_"`
	Symbol_    string        `abi:"symbol_"`
}

func (i *InitializeChildERC721Fn) Sig() []byte {
	return ChildERC721.Abi.Methods["initialize"].ID()
}

func (i *InitializeChildERC721Fn) EncodeAbi() ([]byte, error) {
	return ChildERC721.Abi.Methods["initialize"].Encode(i)
}

func (i *InitializeChildERC721Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721.Abi.Methods["initialize"], buf, i)
}

type OwnerOfChildERC721Fn struct {
	TokenID *big.Int `abi:"tokenId"`
}

func (o *OwnerOfChildERC721Fn) Sig() []byte {
	return ChildERC721.Abi.Methods["ownerOf"].ID()
}

func (o *OwnerOfChildERC721Fn) EncodeAbi() ([]byte, error) {
	return ChildERC721.Abi.Methods["ownerOf"].Encode(o)
}

func (o *OwnerOfChildERC721Fn) DecodeAbi(buf []byte) error {
	return decodeMethod(ChildERC721.Abi.Methods["ownerOf"], buf, o)
}
