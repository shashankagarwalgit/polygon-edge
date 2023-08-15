package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/0xPolygon/polygon-edge/types"
)

// Magic numbers from Ethereum, used in v calculation
var (
	big27 = big.NewInt(27)
	big35 = big.NewInt(35)
)

// TxSigner is a utility interface used to recover data from a transaction
type TxSigner interface {
	// Hash returns the hash of the transaction
	Hash(tx *types.Transaction) types.Hash

	// Sender returns the sender of the transaction
	Sender(tx *types.Transaction) (types.Address, error)

	// SignTx signs a transaction
	SignTx(tx *types.Transaction, priv *ecdsa.PrivateKey) (*types.Transaction, error)
}

// NewSigner creates a new signer object (EIP155 or FrontierSigner)
func NewSigner(forks chain.ForksInTime, chainID uint64) TxSigner {
	var signer TxSigner

	if forks.EIP155 {
		signer = NewEIP155Signer(chainID, forks.Homestead)
	} else {
		signer = NewFrontierSigner(forks.Homestead)
	}

	// London signer requires a fallback signer that is defined above.
	// This is the reason why the london signer check is separated.
	if forks.London {
		return NewLondonSigner(chainID, forks.Homestead, NewBerlinSigner(chainID, forks.Homestead, signer))
	}

	// var eip2930Signer TxSigner
	if forks.EIP2930 {
		return NewBerlinSigner(chainID, forks.Homestead, signer)
	}

	return signer
}

// encodeSignature generates a signature value based on the R, S and V value
func encodeSignature(R, S, V *big.Int, isHomestead bool) ([]byte, error) {
	if !ValidateSignatureValues(V, R, S, isHomestead) {
		return nil, fmt.Errorf("invalid txn signature")
	}

	sig := make([]byte, 65)
	copy(sig[32-len(R.Bytes()):32], R.Bytes())
	copy(sig[64-len(S.Bytes()):64], S.Bytes())
	sig[64] = byte(V.Int64()) // here is safe to convert it since ValidateSignatureValues will validate the v value

	return sig, nil
}

// calcTxHash calculates the transaction hash (keccak256 hash of the RLP value)
func calcTxHash(tx *types.Transaction, chainID uint64) types.Hash {
	var hash []byte

	switch tx.Type() {
	case types.AccessListTx:
		a := signerPool.Get()
		v := a.NewArray()

		v.Set(a.NewUint(chainID))
		v.Set(a.NewUint(tx.Nonce()))
		v.Set(a.NewBigInt(tx.GasPrice()))
		v.Set(a.NewUint(tx.Gas()))

		if tx.To() == nil {
			v.Set(a.NewNull())
		} else {
			v.Set(a.NewCopyBytes((*(tx.To())).Bytes()))
		}

		v.Set(a.NewBigInt(tx.Value()))
		v.Set(a.NewCopyBytes(tx.Input()))

		// add accessList
		accessListVV := a.NewArray()

		if tx.AccessList() != nil {
			for _, accessTuple := range tx.AccessList() {
				accessTupleVV := a.NewArray()
				accessTupleVV.Set(a.NewCopyBytes(accessTuple.Address.Bytes()))

				storageKeysVV := a.NewArray()
				for _, storageKey := range accessTuple.StorageKeys {
					storageKeysVV.Set(a.NewCopyBytes(storageKey.Bytes()))
				}

				accessTupleVV.Set(storageKeysVV)
				accessListVV.Set(accessTupleVV)
			}
		}

		v.Set(accessListVV)

		hash = keccak.PrefixedKeccak256Rlp([]byte{byte(tx.Type())}, nil, v)

		signerPool.Put(a)
	case types.DynamicFeeTx, types.LegacyTx, types.StateTx:
		a := signerPool.Get()
		isDynamicFeeTx := tx.Type() == types.DynamicFeeTx

		v := a.NewArray()

		if isDynamicFeeTx {
			v.Set(a.NewUint(chainID))
		}

		v.Set(a.NewUint(tx.Nonce()))

		if isDynamicFeeTx {
			v.Set(a.NewBigInt(tx.GasTipCap()))
			v.Set(a.NewBigInt(tx.GasFeeCap()))
		} else {
			v.Set(a.NewBigInt(tx.GasPrice()))
		}

		v.Set(a.NewUint(tx.Gas()))

		if tx.To() == nil {
			v.Set(a.NewNull())
		} else {
			v.Set(a.NewCopyBytes((*(tx.To())).Bytes()))
		}

		v.Set(a.NewBigInt(tx.Value()))

		v.Set(a.NewCopyBytes(tx.Input()))

		if isDynamicFeeTx {
			//v.Set(a.NewArray())
			// Convert TxAccessList to RLP format and add it to the vv array.
			accessListVV := a.NewArray()

			if tx.AccessList() != nil {
				for _, accessTuple := range tx.AccessList() {
					accessTupleVV := a.NewArray()
					accessTupleVV.Set(a.NewCopyBytes(accessTuple.Address.Bytes()))

					storageKeysVV := a.NewArray()
					for _, storageKey := range accessTuple.StorageKeys {
						storageKeysVV.Set(a.NewCopyBytes(storageKey.Bytes()))
					}

					accessTupleVV.Set(storageKeysVV)
					accessListVV.Set(accessTupleVV)
				}
			}

			v.Set(accessListVV)
		} else {
			// EIP155
			if chainID != 0 {
				v.Set(a.NewUint(chainID))
				v.Set(a.NewUint(0))
				v.Set(a.NewUint(0))
			}
		}

		if isDynamicFeeTx {
			hash = keccak.PrefixedKeccak256Rlp([]byte{byte(tx.Type())}, nil, v)
		} else {
			hash = keccak.Keccak256Rlp(nil, v)
		}

		signerPool.Put(a)
	}

	return types.BytesToHash(hash)
}
