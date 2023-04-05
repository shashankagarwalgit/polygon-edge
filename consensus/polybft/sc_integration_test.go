package polybft

import (
	"encoding/hex"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/bitmap"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/contractsapi"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/contractsapi/artifact"
	bls "github.com/0xPolygon/polygon-edge/consensus/polybft/signer"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/wallet"
	"github.com/0xPolygon/polygon-edge/contracts"
	"github.com/0xPolygon/polygon-edge/crypto"
	secretsHelper "github.com/0xPolygon/polygon-edge/secrets/helper"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

func TestIntegratoin_PerformExit(t *testing.T) {
	t.Parallel()

	const gasPrice = 1000000000000

	// create validator set and checkpoint mngr
	currentValidators := newTestValidatorsWithAliases(t, []string{"A", "B", "C", "D"}, []uint64{100, 100, 100, 100})
	accSet := currentValidators.getPublicIdentities()
	cm := checkpointManager{blockchain: &blockchainMock{}}

	senderAddress := types.Address{1}
	bn256Addr := types.Address{2}
	rootERC20Token := types.Address{4}
	stateSenderAddr := types.Address{5}
	receiverAddr := types.Address{6}
	amount1 := big.NewInt(3)
	amount2 := big.NewInt(2)

	alloc := map[types.Address]*chain.GenesisAccount{
		senderAddress:         {Balance: big.NewInt(100000000000)},
		contracts.BLSContract: {Code: contractsapi.BLS.DeployedBytecode},
		bn256Addr:             {Code: contractsapi.BLS256.DeployedBytecode},
		rootERC20Token:        {Code: contractsapi.RootERC20.DeployedBytecode},
		stateSenderAddr:       {Code: contractsapi.StateSender.DeployedBytecode},
	}
	transition := newTestTransition(t, alloc)

	getField := func(addr types.Address, abi *abi.ABI, function string, args ...interface{}) []byte {
		input, err := abi.GetMethod(function).Encode(args)
		require.NoError(t, err)

		result := transition.Call2(senderAddress, addr, input, big.NewInt(0), 1000000000)
		require.NoError(t, result.Err)
		require.True(t, result.Succeeded())
		require.False(t, result.Failed())

		return result.ReturnValue
	}

	// deploy CheckpointManager
	checkpointManagerInit := func() ([]byte, error) {
		return (&contractsapi.InitializeCheckpointManagerFn{
			NewBls:          contracts.BLSContract,
			NewBn256G2:      bn256Addr,
			NewValidatorSet: accSet.ToAPIBinding(),
			ChainID_:        big.NewInt(0),
		}).EncodeAbi()
	}

	checkpointManagerAddr := deployAndInitContract(t, transition, contractsapi.CheckpointManager, senderAddress, checkpointManagerInit)

	// deploy ExitHelper
	exitHelperInit := func() ([]byte, error) {
		return (&contractsapi.InitializeExitHelperFn{NewCheckpointManager: checkpointManagerAddr}).EncodeAbi()
	}
	exitHelperContractAddress := deployAndInitContract(t, transition, contractsapi.ExitHelper, senderAddress, exitHelperInit)

	// deploy RootERC20Predicate
	rootERC20PredicateInit := func() ([]byte, error) {
		return (&contractsapi.InitializeRootERC20PredicateFn{
			NewStateSender:         stateSenderAddr,
			NewExitHelper:          exitHelperContractAddress,
			NewChildERC20Predicate: contracts.ChildERC20PredicateContract,
			NewChildTokenTemplate:  contracts.ChildERC20Contract,
			NativeTokenRootAddress: contracts.NativeERC20TokenContract,
		}).EncodeAbi()
	}
	rootERC20PredicateAddr := deployAndInitContract(t, transition, contractsapi.RootERC20Predicate, senderAddress, rootERC20PredicateInit)

	require.Equal(t, getField(checkpointManagerAddr, contractsapi.CheckpointManager.Abi, "currentCheckpointBlockNumber")[31], uint8(0))

	accSetHash, err := accSet.Hash()
	require.NoError(t, err)

	blockHash := types.Hash{5}
	blockNumber := uint64(1)
	epochNumber := uint64(1)
	blockRound := uint64(1)

	// deposit
	depositFn, err := (&contractsapi.DepositToRootERC20PredicateFn{
		RootToken: rootERC20Token,
		Receiver:  receiverAddr,
		Amount:    new(big.Int).Add(amount1, amount2),
	}).EncodeAbi()
	require.NoError(t, err)

	result := transition.Call2(senderAddress, rootERC20PredicateAddr, depositFn, big.NewInt(0), 1000000000)
	require.NoError(t, result.Err)

	widthdrawSig := crypto.Keccak256([]byte("WITHDRAW"))
	erc20DataType := abi.MustNewType(
		"tuple(bytes32 withdrawSignature, address rootToken, address withdrawer, address receiver, uint256 amount)")

	exitData1, err := erc20DataType.Encode(map[string]interface{}{
		"withdrawSignature": widthdrawSig,
		"rootToken":         ethgo.Address(rootERC20Token),
		"withdrawer":        ethgo.Address(senderAddress),
		"receiver":          ethgo.Address(receiverAddr),
		"amount":            amount1,
	})
	require.NoError(t, err)

	exitData2, err := erc20DataType.Encode(map[string]interface{}{
		"withdrawSignature": widthdrawSig,
		"rootToken":         ethgo.Address(rootERC20Token),
		"withdrawer":        ethgo.Address(senderAddress),
		"receiver":          ethgo.Address(receiverAddr),
		"amount":            amount2,
	})
	require.NoError(t, err)

	exits := []*ExitEvent{
		{
			ID:       1,
			Sender:   ethgo.Address(contracts.ChildERC20PredicateContract),
			Receiver: ethgo.Address(rootERC20PredicateAddr),
			Data:     exitData1,
		},
		{
			ID:       2,
			Sender:   ethgo.Address(contracts.ChildERC20PredicateContract),
			Receiver: ethgo.Address(rootERC20PredicateAddr),
			Data:     exitData2,
		},
	}
	exitTree, err := createExitTree(exits)
	require.NoError(t, err)

	eventRoot := exitTree.Hash()

	checkpointData := CheckpointData{
		BlockRound:            blockRound,
		EpochNumber:           epochNumber,
		CurrentValidatorsHash: accSetHash,
		NextValidatorsHash:    accSetHash,
		EventRoot:             eventRoot,
	}

	checkpointHash, err := checkpointData.Hash(
		cm.blockchain.GetChainID(),
		blockRound,
		blockHash)
	require.NoError(t, err)

	i := uint64(0)
	bmp := bitmap.Bitmap{}
	signatures := bls.Signatures(nil)

	currentValidators.iterAcct(nil, func(v *testValidator) {
		signatures = append(signatures, v.mustSign(checkpointHash[:], bls.DomainCheckpointManager))
		bmp.Set(i)
		i++
	})

	aggSignature, err := signatures.Aggregate().Marshal()
	require.NoError(t, err)

	extra := &Extra{
		Checkpoint: &checkpointData,
		Committed: &Signature{
			AggregatedSignature: aggSignature,
			Bitmap:              bmp,
		},
	}

	// submit a checkpoint
	submitCheckpointEncoded, err := cm.abiEncodeCheckpointBlock(blockNumber, blockHash, extra, accSet)
	require.NoError(t, err)

	result = transition.Call2(senderAddress, checkpointManagerAddr, submitCheckpointEncoded, big.NewInt(0), 1000000000)
	require.NoError(t, result.Err)
	require.Equal(t, getField(checkpointManagerAddr, contractsapi.CheckpointManager.Abi, "currentCheckpointBlockNumber")[31], uint8(1))

	// check that the exit hasn't performed
	res := getField(exitHelperContractAddress, contractsapi.ExitHelper.Abi, "processedExits", exits[0].ID)
	require.Equal(t, int(res[31]), 0)

	var exitEventAPI contractsapi.L2StateSyncedEvent
	proofExitEvent, err := exitEventAPI.Encode(exits[0])
	require.NoError(t, err)

	proof, err := exitTree.GenerateProof(proofExitEvent)
	require.NoError(t, err)

	leafIndex, err := exitTree.LeafIndex(proofExitEvent)
	require.NoError(t, err)

	exitFnInput, err := (&contractsapi.ExitExitHelperFn{
		BlockNumber:  new(big.Int).SetUint64(blockNumber),
		LeafIndex:    new(big.Int).SetUint64(leafIndex),
		UnhashedLeaf: proofExitEvent,
		Proof:        proof,
	}).EncodeAbi()
	require.NoError(t, err)

	result = transition.Call2(senderAddress, exitHelperContractAddress, exitFnInput, big.NewInt(0), 1000000000)
	require.NoError(t, result.Err)

	// check true
	res = getField(exitHelperContractAddress, contractsapi.ExitHelper.Abi, "processedExits", exits[0].ID)
	require.Equal(t, types.ReceiptSuccess, int(res[31]))

	lastID := getField(rootERC20Token, contractsapi.RootERC20.Abi, "id")
	require.Equal(t, uint8(1), lastID[31])

	lastAddr := getField(rootERC20Token, contractsapi.RootERC20.Abi, "addr")
	require.Equal(t, exits[0].Sender[:], lastAddr[12:])

	lastCounter := getField(rootERC20Token, contractsapi.RootERC20.Abi, "counter")
	require.Equal(t, uint8(1), lastCounter[31])
}

func TestIntegration_CommitEpoch(t *testing.T) {
	t.Parallel()

	// init validator sets
	// (cannot run test case with more than 100 validators at the moment,
	// because active validator set is capped to 100 on smart contract side)
	validatorSetSize := []int{5, 10, 50, 100}
	// number of delegators per validator
	delegatorsPerValidator := 100

	intialBalance := uint64(5 * math.Pow(10, 18))  // 5 tokens
	reward := uint64(math.Pow(10, 18))             // 1 token
	delegateAmount := uint64(math.Pow(10, 18)) / 2 // 0.5 token

	validatorSets := make([]*testValidators, len(validatorSetSize), len(validatorSetSize))

	// create all validator sets which will be used in test
	for i, size := range validatorSetSize {
		aliases := make([]string, size, size)
		vps := make([]uint64, size, size)

		for j := 0; j < size; j++ {
			aliases[j] = "v" + strconv.Itoa(j)
			vps[j] = intialBalance
		}

		validatorSets[i] = newTestValidatorsWithAliases(t, aliases, vps)
	}

	// iterate through the validator set and do the test for each of them
	for _, currentValidators := range validatorSets {
		accSet := currentValidators.getPublicIdentities()
		accSetPrivateKeys := currentValidators.getPrivateIdentities()
		valid2deleg := make(map[types.Address][]*wallet.Key, accSet.Len()) // delegators assigned to validators

		// add contracts to genesis data
		alloc := map[types.Address]*chain.GenesisAccount{
			contracts.ValidatorSetContract: {
				Code: contractsapi.ChildValidatorSet.DeployedBytecode,
			},
			contracts.BLSContract: {
				Code: contractsapi.BLS.DeployedBytecode,
			},
		}

		// validator data for polybft config
		initValidators := make([]*Validator, accSet.Len())

		for i, validator := range accSet {
			// add validator to genesis data
			alloc[validator.Address] = &chain.GenesisAccount{
				Balance: validator.VotingPower,
			}

			signature, err := secretsHelper.MakeKOSKSignature(accSetPrivateKeys[i].Bls, validator.Address, 0, bls.DomainValidatorSet)
			require.NoError(t, err)

			signatureBytes, err := signature.Marshal()
			require.NoError(t, err)

			// create validator data for polybft config
			initValidators[i] = &Validator{
				Address:      validator.Address,
				Balance:      validator.VotingPower,
				Stake:        validator.VotingPower,
				BlsKey:       hex.EncodeToString(validator.BlsKey.Marshal()),
				BlsSignature: hex.EncodeToString(signatureBytes),
			}

			// create delegators
			delegatorAccs := createRandomTestKeys(t, delegatorsPerValidator)

			// add delegators to genesis data
			for j := 0; j < delegatorsPerValidator; j++ {
				delegator := delegatorAccs[j]
				alloc[types.Address(delegator.Address())] = &chain.GenesisAccount{
					Balance: new(big.Int).SetUint64(intialBalance),
				}
			}

			valid2deleg[validator.Address] = delegatorAccs
		}

		transition := newTestTransition(t, alloc)

		polyBFTConfig := PolyBFTConfig{
			InitialValidatorSet: initValidators,
			EpochSize:           24 * 60 * 60 / 2,
			SprintSize:          5,
			EpochReward:         reward,
			// use 1st account as governance address
			Governance: currentValidators.toValidatorSet().validators.GetAddresses()[0],
		}

		// get data for ChildValidatorSet initialization
		initInput, err := getInitChildValidatorSetInput(polyBFTConfig)
		require.NoError(t, err)

		// init ChildValidatorSet
		err = initContract(contracts.ValidatorSetContract, initInput, "ChildValidatorSet", transition)
		require.NoError(t, err)

		// delegate amounts to validators
		for valAddress, delegators := range valid2deleg {
			for _, delegator := range delegators {
				encoded, err := contractsapi.ChildValidatorSet.Abi.Methods["delegate"].Encode(
					[]interface{}{valAddress, false})
				require.NoError(t, err)

				result := transition.Call2(types.Address(delegator.Address()), contracts.ValidatorSetContract, encoded, new(big.Int).SetUint64(delegateAmount), 1000000000000)
				require.False(t, result.Failed())
			}
		}

		// create input for commit epoch
		commitEpoch := createTestCommitEpochInput(t, 1, accSet, polyBFTConfig.EpochSize)
		input, err := commitEpoch.EncodeAbi()
		require.NoError(t, err)

		// call commit epoch
		result := transition.Call2(contracts.SystemCaller, contracts.ValidatorSetContract, input, big.NewInt(0), 10000000000)
		require.NoError(t, result.Err)
		t.Logf("Number of validators %d when we add %d of delegators, Gas used %+v\n", accSet.Len(), accSet.Len()*delegatorsPerValidator, result.GasUsed)

		commitEpoch = createTestCommitEpochInput(t, 2, accSet, polyBFTConfig.EpochSize)
		input, err = commitEpoch.EncodeAbi()
		require.NoError(t, err)

		// call commit epoch
		result = transition.Call2(contracts.SystemCaller, contracts.ValidatorSetContract, input, big.NewInt(0), 10000000000)
		require.NoError(t, result.Err)
		t.Logf("Number of validators %d, Number of delegator %d, Gas used %+v\n", accSet.Len(), accSet.Len()*delegatorsPerValidator, result.GasUsed)
	}
}

func deployAndInitContract(t *testing.T, transition *state.Transition, scArtifact *artifact.Artifact, sender types.Address,
	initCallback func() ([]byte, error)) types.Address {
	t.Helper()

	deployResult := transition.Create2(sender, scArtifact.Bytecode, big.NewInt(0), 1e9)
	assert.NoError(t, deployResult.Err)

	if initCallback != nil {
		initInput, err := initCallback()
		require.NoError(t, err)

		result := transition.Call2(sender, deployResult.Address, initInput, big.NewInt(0), 1e9)
		require.NoError(t, result.Err)
	}

	return deployResult.Address
}
