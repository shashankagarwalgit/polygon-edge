package polybft

import (
	"bytes"
	"math"
	"testing"

	bls "github.com/0xPolygon/polygon-edge/consensus/polybft/signer"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValSetIndex(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(5)
	require.NoError(t, err)

	addresses := []types.Address{{0x10}, {0x52}, {0x33}, {0x74}, {0x60}}

	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     addresses[0],
			VotingPower: 10,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     addresses[1],
			VotingPower: 100,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     addresses[2],
			VotingPower: 1,
		},
		{
			BlsKey:      keys[3].PublicKey(),
			Address:     addresses[3],
			VotingPower: 50,
		},
		{
			BlsKey:      keys[4].PublicKey(),
			Address:     addresses[4],
			VotingPower: 30,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)
	// validate no changes to validator set positions
	for i, v := range vs.Accounts() {
		assert.Equal(t, addresses[i], v.Address)
	}

	proposer, err := vs.CalcProposer(0)
	require.NoError(t, err)
	assert.Equal(t, proposer, addresses[1])

	// validate no changes to validator set positions
	for i, v := range vs.Accounts() {
		assert.Equal(t, addresses[i], v.Address)
	}
}

func TestCalculateProposer(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(5)
	require.NoError(t, err)

	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 2,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 3,
		},
		{
			BlsKey:      keys[3].PublicKey(),
			Address:     types.Address{0x4},
			VotingPower: 4,
		},
		{
			BlsKey:      keys[4].PublicKey(),
			Address:     types.Address{0x5},
			VotingPower: 5,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)
	assert.Equal(t, int64(15), vs.totalVotingPower)

	currProposerAddress, err := vs.CalcProposer(0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x5}, currProposerAddress)

	proposerAddressR1, err := vs.CalcProposer(1)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x4}, proposerAddressR1)

	proposerAddressR2, err := vs.CalcProposer(2)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x3}, proposerAddressR2)

	proposerAddressR3, err := vs.CalcProposer(3)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerAddressR3)

	proposerAddressR4, err := vs.CalcProposer(4)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x5}, proposerAddressR4)

	proposerAddressR5, err := vs.CalcProposer(5)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x4}, proposerAddressR5)

	proposerAddressR6, err := vs.CalcProposer(6)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerAddressR6)
}

func TestCalcProposer(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(5)
	require.NoError(t, err)

	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 2,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 3,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)

	proposerR0, err := vs.CalcProposer(0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x3}, proposerR0)

	proposerR1, err := vs.CalcProposer(1)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerR1)

	proposerR2, err := vs.CalcProposer(2)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerR2)
}

func TestProposerSelection1(t *testing.T) {
	t.Parallel()

	const numberOfIteration = 99

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	vset, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1000,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 300,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 330,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)

	var proposers = make([]types.Address, numberOfIteration)

	for i := uint64(0); i < numberOfIteration; i++ {
		proposers[i], err = vset.CalcProposer(i)
		require.NoError(t, err)
	}

	expected := []types.Address{
		{0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1},
		{0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1},
		{0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3},
		{0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x1}, {0x3}, {0x2}, {0x1}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2},
		{0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1}, {0x1}, {0x3}, {0x1}, {0x2}, {0x1},
		{0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1}, {0x3}, {0x1}, {0x1}, {0x2}, {0x1},
		{0x3}, {0x1}, {0x1},
	}

	for i, p := range proposers {
		assert.True(t, bytes.Equal(expected[i].Bytes(), p.Bytes()))
	}
}

// Test that IncrementProposerPriority requires positive times.
func TestIncrementProposerPriorityPositiveTimes(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	vset, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1000,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 300,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 330,
		},
	}, hclog.NewNullLogger())

	require.NoError(t, err)
	proposerAddressR0, err := vset.CalcProposer(0)
	assert.Equal(t, types.Address{0x1}, proposerAddressR0)

	// priority must be > 0
	err = vset.IncrementProposerPriority(0)
	require.Error(t, err)

	proposerAddressR1, err := vset.CalcProposer(1)
	require.NoError(t, err)

	assert.Equal(t, types.Address{0x3}, proposerAddressR1)
}

func TestIncrementProposerPrioritySameVotingPower(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	vs, err := NewValidatorSet([]*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: 1,
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: 1,
		},
	}, hclog.NewNullLogger())
	require.NoError(t, err)
	assert.Equal(t, int64(3), vs.totalVotingPower)

	// when voting power is the same order is by address
	currProposerAddress, err := vs.CalcProposer(0)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, currProposerAddress)

	proposerAddresR1, err := vs.CalcProposer(1)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerAddresR1)

	proposerAddressR2, err := vs.CalcProposer(2)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x3}, proposerAddressR2)

	proposerAddressR3, err := vs.CalcProposer(3)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x1}, proposerAddressR3)

	proposerAddressR4, err := vs.CalcProposer(4)
	require.NoError(t, err)
	assert.Equal(t, types.Address{0x2}, proposerAddressR4)
}

func TestAveragingInIncrementProposerPriorityWithVotingPower(t *testing.T) {
	t.Parallel()

	keys, err := bls.CreateRandomBlsKeys(3)
	require.NoError(t, err)

	// Other than TestAveragingInIncrementProposerPriority this is a more complete test showing
	// how each ProposerPriority changes in relation to the validator's voting power respectively.
	// average is zero in each round:
	vp0 := int64(10)
	vp1 := int64(1)
	vp2 := int64(1)
	total := vp0 + vp1 + vp2
	avg := (vp0 + vp1 + vp2 - total) / 3
	valz := []*ValidatorMetadata{
		{
			BlsKey:      keys[0].PublicKey(),
			Address:     types.Address{0x1},
			VotingPower: uint64(vp0),
		},
		{
			BlsKey:      keys[1].PublicKey(),
			Address:     types.Address{0x2},
			VotingPower: uint64(vp1),
		},
		{
			BlsKey:      keys[2].PublicKey(),
			Address:     types.Address{0x3},
			VotingPower: uint64(vp2),
		},
	}

	vals, err := NewValidatorSet(valz, hclog.NewNullLogger())
	assert.NoError(t, err)

	tcs := []struct {
		vals                  *validatorSet
		wantProposerPrioritys []int64
		times                 uint64
		wantProposer          *ValidatorMetadata
	}{

		0: {
			vals.Copy(),
			[]int64{
				// Acumm+VotingPower-Avg:
				0 + vp0 - total - avg, // mostest will be subtracted by total voting power (12)
				0 + vp1,
				0 + vp2},
			1,
			vals.validators[0].Metadata,
		},
		1: {
			vals.Copy(),
			[]int64{
				(0 + vp0 - total) + vp0 - total - avg, // this will be mostest on 2nd iter, too
				(0 + vp1) + vp1,
				(0 + vp2) + vp2},
			2,
			vals.validators[0].Metadata,
		}, // increment twice -> expect average to be subtracted twice
		2: {
			vals.Copy(),
			[]int64{
				0 + 3*(vp0-total) - avg, // still mostest
				0 + 3*vp1,
				0 + 3*vp2},
			3,
			vals.validators[0].Metadata,
		},
		3: {
			vals.Copy(),
			[]int64{
				0 + 4*(vp0-total), // still mostest
				0 + 4*vp1,
				0 + 4*vp2},
			4,
			vals.validators[0].Metadata,
		},
		4: {
			vals.Copy(),
			[]int64{
				0 + 4*(vp0-total) + vp0, // 4 iters was mostest
				0 + 5*vp1 - total,       // now this val is mostest for the 1st time (hence -12==totalVotingPower)
				0 + 5*vp2},
			5,
			vals.validators[1].Metadata,
		},
		5: {
			vals.Copy(),
			[]int64{
				0 + 6*vp0 - 5*total, // mostest again
				0 + 6*vp1 - total,   // mostest once up to here
				0 + 6*vp2},
			6,
			vals.validators[0].Metadata,
		},
		6: {
			vals.Copy(),
			[]int64{
				0 + 7*vp0 - 6*total, // in 7 iters this val is mostest 6 times
				0 + 7*vp1 - total,   // in 7 iters this val is mostest 1 time
				0 + 7*vp2},
			7,
			vals.validators[0].Metadata,
		},
		7: {
			vals.Copy(),
			[]int64{
				0 + 8*vp0 - 7*total, // mostest again
				0 + 8*vp1 - total,
				0 + 8*vp2},
			8,
			vals.validators[0].Metadata,
		},
		8: {
			vals.Copy(),
			[]int64{
				0 + 9*vp0 - 7*total,
				0 + 9*vp1 - total,
				0 + 9*vp2 - total}, // mostest
			9,
			vals.validators[2].Metadata,
		},
		9: {
			vals.Copy(),
			[]int64{
				0 + 10*vp0 - 8*total, // after 10 iters this is mostest again
				0 + 10*vp1 - total,   // after 6 iters this val is "mostest" once and not in between
				0 + 10*vp2 - total},  // in between 10 iters this val is "mostest" once
			10,
			vals.validators[0].Metadata,
		},
		10: {
			vals.Copy(),
			[]int64{
				0 + 11*vp0 - 9*total,
				0 + 11*vp1 - total,  // after 6 iters this val is "mostest" once and not in between
				0 + 11*vp2 - total}, // after 10 iters this val is "mostest" once
			11,
			vals.validators[0].Metadata,
		},
	}

	for i, tc := range tcs {
		err := tc.vals.IncrementProposerPriority(tc.times)
		assert.NoError(t, err)

		proposer, err := tc.vals.getProposer()
		assert.NoError(t, err)
		assert.Equal(t, tc.wantProposer.Address, proposer.Metadata.Address,
			"test case: %v",
			i)

		for valIdx, val := range tc.vals.validators {
			assert.Equal(t,
				tc.wantProposerPrioritys[valIdx],
				val.ProposerPriority,
				"test case: %v, validator: %v",
				i,
				valIdx)
		}
	}
}

func TestValidatorSetTotalVotingPowerErrorOnOverflow(t *testing.T) {
	t.Parallel()

	// NewValidatorSet calls IncrementProposerPriority which calls TotalVotingPower()
	// which should panic on overflows:
	_, err := NewValidatorSet([]*ValidatorMetadata{
		{Address: types.Address{0x1}, VotingPower: math.MaxInt64},
		{Address: types.Address{0x2}, VotingPower: math.MaxInt64},
		{Address: types.Address{0x3}, VotingPower: math.MaxInt64},
	}, hclog.NewNullLogger())
	require.Error(t, err)
}

func TestUpdatesForNewValidatorSet(t *testing.T) {
	t.Parallel()

	v1 := &ValidatorMetadata{Address: types.Address{0x1}, VotingPower: 100}
	v2 := &ValidatorMetadata{Address: types.Address{0x2}, VotingPower: 100}
	accountSet := []*ValidatorMetadata{v1, v2}
	valSet, err := NewValidatorSet(accountSet, hclog.NewNullLogger())
	require.NoError(t, err)

	_, err = valSet.CalcProposer(1)
	require.NoError(t, err)
	verifyValidatorSet(t, valSet)
}

func TestValidatorSet_HasQuorum(t *testing.T) {
	t.Parallel()

	t.Run("BFT conditions met", func(t *testing.T) {
		t.Parallel()
		// enough signers for quorum (2/3 super-majority of validators are signers)
		validators := newTestValidatorsWithAliases([]string{"A", "B", "C", "D", "E", "F", "G"})
		vs, err := validators.toValidatorSet()
		require.NoError(t, err)

		signers := []types.Address{}
		validators.iterAcct([]string{"A", "B", "C", "D", "E"}, func(v *testValidator) {
			signers = append(signers, v.Address())
		})

		require.True(t, vs.HasQuorum(signers))

		// not enough signers for quorum (less than 2/3 super-majority of validators are signers)
		signers = []types.Address{}
		validators.iterAcct([]string{"A", "B", "C", "D"}, func(v *testValidator) {
			signers = append(signers, v.Address())
		})
		require.False(t, vs.HasQuorum(signers))
	})

	t.Run("BFT conditions aren't met", func(t *testing.T) {
		t.Parallel()
		validators := newTestValidatorsWithAliases([]string{"A", "B", "C"})
		vs, err := validators.toValidatorSet()
		require.NoError(t, err)

		// BFT conditions aren't met (less than 5 validators)
		// in order to have a quorum satisfied, all the validators need to be among signers
		signers := []types.Address{}
		validators.iterAcct([]string{"A", "B", "C"}, func(v *testValidator) {
			signers = append(signers, v.Address())
		})
		require.True(t, vs.HasQuorum(signers))

		// not entire validator set are among signers, quorum isn't satisfied
		signers = []types.Address{}
		validators.iterAcct([]string{"A", "B"}, func(v *testValidator) {
			signers = append(signers, v.Address())
		})
		require.False(t, vs.HasQuorum(signers))
	})
}

func TestValidatorSet_HasQuorumWithoutProposer(t *testing.T) {
	t.Parallel()

	validators := newTestValidatorsWithAliases([]string{"A", "B", "C"})
	vs, err := validators.toValidatorSet()
	require.NoError(t, err)

	_, err = vs.CalcProposer(0)
	require.NoError(t, err)

	// BFT conditions aren't met (less than 5 validators)
	// in order to have a quorum satisfied, all the validators need to be among signers
	signers := []types.Address{}

	validators.iterAcct([]string{"B", "C"}, func(v *testValidator) {
		signers = append(signers, v.Address())
	})
	require.True(t, vs.HasQuorumWithoutProposer(signers))

	// no quorum, since only a single validator is signed
	// (it doesn't have enough voting power, even when proposer voting power is subtracted from quorum size)
	signers = []types.Address{}

	validators.iterAcct([]string{"B"}, func(v *testValidator) {
		signers = append(signers, v.Address())
	})
	require.False(t, vs.HasQuorumWithoutProposer(signers))
}

func verifyValidatorSet(t *testing.T, valSet *validatorSet) {
	t.Helper()
	// verify that the capacity and length of validators is the same
	assert.Equal(t, len(valSet.Accounts()), cap(valSet.validators))
	// verify that the set's total voting power has been updated
	tvp := valSet.totalVotingPower
	err := valSet.updateTotalVotingPower()
	require.NoError(t, err)
	expectedTvp, err := valSet.TotalVotingPower()
	require.NoError(t, err)
	assert.Equal(t, expectedTvp, tvp,
		"expected TVP %d. Got %d, valSet=%s", expectedTvp, tvp, valSet)
	// verify that validator priorities are centered
	valsCount := int64(len(valSet.validators))
	tpp := valSetTotalProposerPriority(valSet)
	assert.True(t, tpp < valsCount && tpp > -valsCount,
		"expected total priority in (-%d, %d). Got %d", valsCount, valsCount, tpp)
	// verify that priorities are scaled
	dist := computeMaxMinPriorityDiff(valSet)
	assert.True(t, dist <= priorityWindowSizeFactor*tvp,
		"expected priority distance < %d. Got %d", priorityWindowSizeFactor*tvp, dist)
}

func valSetTotalProposerPriority(valSet *validatorSet) int64 {
	sum := int64(0)
	for _, val := range valSet.validators {
		// mind overflow
		sum = safeAddClip(sum, val.ProposerPriority)
	}

	return sum
}
