package polybft

import (
	"testing"

	"github.com/0xPolygon/polygon-edge/consensus/polybft/bitmap"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/wallet"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/stretchr/testify/assert"
)

func Test_setupHeaderHashFunc(t *testing.T) {
	extra := &Extra{
		Validators: &ValidatorSetDelta{Removed: bitmap.Bitmap{1}},
		Parent:     createSignature(t, []*wallet.Account{wallet.GenerateAccount()}, types.ZeroHash),
		Checkpoint: &CheckpointData{},
	}

	header := &types.Header{
		Number:    2,
		GasLimit:  10000003,
		Timestamp: 18,
	}

	extra.Seal = []byte{}
	extra.Committed = &Signature{}
	header.ExtraData = extra.MarshalRLPTo(nil)
	oldHash := types.HeaderHash(header)

	extra.Seal = []byte{1, 2, 3, 255}
	extra.Committed = createSignature(t, []*wallet.Account{wallet.GenerateAccount()}, types.ZeroHash)
	header.ExtraData = append(make([]byte, ExtraVanity), extra.MarshalRLPTo(nil)...)
	oldHashWithSealCommited := types.HeaderHash(header)

	setupHeaderHashFunc()

	result := types.HeaderHash(header)

	assert.NotEqual(t, oldHashWithSealCommited, result)
	assert.Equal(t, oldHash, result)

	header.ExtraData = []byte{1, 2, 3, 4, 100, 200, 255}
	assert.Equal(t, types.ZeroHash, types.HeaderHash(header)) // to small extra data
}
