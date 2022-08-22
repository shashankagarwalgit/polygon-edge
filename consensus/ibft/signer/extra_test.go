package signer

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/0xPolygon/polygon-edge/validators"
	"github.com/stretchr/testify/assert"
)

var (
	validatorAddr1         = types.StringToAddress("1")
	validatorBLSPublicKey1 = validators.BLSValidatorPublicKey([]byte{0x1})

	testProposerSeal = []byte{0x1}
)

func JsonMarshalHelper(t *testing.T, extra *IstanbulExtra) string {
	t.Helper()

	res, err := json.Marshal(extra)

	assert.NoError(t, err)

	return string(res)
}

func TestIstanbulExtraMarshalAndUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		extra *IstanbulExtra
	}{
		{
			name: "ECDSAExtra",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
				ParentCommittedSeals: &SerializedSeal{
					[]byte{0x3},
					[]byte{0x4},
				},
			},
		},
		{
			name: "ECDSAExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
			},
		},
		{
			name: "BLSExtra",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
		},
		{
			name: "BLSExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create original data
			originalExtraJson := JsonMarshalHelper(t, test.extra)

			bytesData := test.extra.MarshalRLPTo(nil)
			err := test.extra.UnmarshalRLP(bytesData)
			assert.NoError(t, err)

			// make sure all data is recovered
			assert.Equal(
				t,
				originalExtraJson,
				JsonMarshalHelper(t, test.extra),
			)
		})
	}
}

func Test_packProposerSealIntoExtra(t *testing.T) {
	newProposerSeal := []byte("new proposer seal")

	tests := []struct {
		name  string
		extra *IstanbulExtra
	}{
		{
			name: "ECDSAExtra",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
				ParentCommittedSeals: &SerializedSeal{
					[]byte{0x3},
					[]byte{0x4},
				},
			},
		},
		{
			name: "ECDSAExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
			},
		},
		{
			name: "BLSExtra",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
		},
		{
			name: "BLSExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalProposerSeal := test.extra.ProposerSeal

			// create expected data
			test.extra.ProposerSeal = newProposerSeal
			expectedJSON := JsonMarshalHelper(t, test.extra)
			test.extra.ProposerSeal = originalProposerSeal

			newExtraBytes := packProposerSealIntoExtra(
				// prepend IstanbulExtraHeader to parse
				append(
					make([]byte, IstanbulExtraVanity),
					test.extra.MarshalRLPTo(nil)...,
				),
				newProposerSeal,
			)

			assert.NoError(
				t,
				test.extra.UnmarshalRLP(newExtraBytes[IstanbulExtraVanity:]),
			)

			// check json of decoded data matches with the original data
			jsonData := JsonMarshalHelper(t, test.extra)

			assert.Equal(
				t,
				expectedJSON,
				jsonData,
			)
		})
	}
}

func Test_packCommittedSealsIntoExtra(t *testing.T) {
	tests := []struct {
		name              string
		extra             *IstanbulExtra
		newCommittedSeals Sealer
	}{
		{
			name: "ECDSAExtra",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
				ParentCommittedSeals: &SerializedSeal{
					[]byte{0x3},
					[]byte{0x4},
				},
			},
			newCommittedSeals: &SerializedSeal{
				[]byte{0x3},
				[]byte{0x4},
			},
		},
		{
			name: "ECDSAExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.ECDSAValidators{
					&validators.ECDSAValidator{
						Address: validatorAddr1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &SerializedSeal{
					[]byte{0x1},
					[]byte{0x2},
				},
			},
			newCommittedSeals: &SerializedSeal{
				[]byte{0x3},
				[]byte{0x4},
			},
		},
		{
			name: "BLSExtra",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
			newCommittedSeals: &BLSSeal{
				Bitmap:    new(big.Int).SetBytes([]byte{0xa}),
				Signature: []byte{0x2},
			},
		},
		{
			name: "BLSExtra without ParentCommittedSeals",
			extra: &IstanbulExtra{
				Validators: &validators.BLSValidators{
					&validators.BLSValidator{
						Address:      validatorAddr1,
						BLSPublicKey: validatorBLSPublicKey1,
					},
				},
				ProposerSeal: testProposerSeal,
				CommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x8}),
					Signature: []byte{0x1},
				},
				ParentCommittedSeals: &BLSSeal{
					Bitmap:    new(big.Int).SetBytes([]byte{0x9}),
					Signature: []byte{0x2},
				},
			},
			newCommittedSeals: &BLSSeal{
				Bitmap:    new(big.Int).SetBytes([]byte{0xa}),
				Signature: []byte{0x2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalCommittedSeals := test.extra.CommittedSeals

			// create expected data
			test.extra.CommittedSeals = test.newCommittedSeals
			expectedJSON := JsonMarshalHelper(t, test.extra)
			test.extra.CommittedSeals = originalCommittedSeals

			// update committed seals
			newExtraBytes := packCommittedSealsIntoExtra(
				// prepend IstanbulExtraHeader
				append(
					make([]byte, IstanbulExtraVanity),
					test.extra.MarshalRLPTo(nil)...,
				),
				test.newCommittedSeals,
			)

			// decode RLP data
			assert.NoError(
				t,
				test.extra.UnmarshalRLP(newExtraBytes[IstanbulExtraVanity:]),
			)

			// check json of decoded data matches with the original data
			jsonData := JsonMarshalHelper(t, test.extra)

			assert.Equal(
				t,
				expectedJSON,
				jsonData,
			)
		})
	}
}
