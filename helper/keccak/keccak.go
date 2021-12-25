package keccak

import (
	"hash"

	"github.com/umbracle/fastrlp"
	"golang.org/x/crypto/sha3"
)

type hashImpl interface {
	hash.Hash
	Read(b []byte) (int, error)
}

// Keccak is the sha256 keccak hash
type Keccak struct {
	buf  []byte // buffer to store intermediate rlp marshal values
	tmp  []byte
	hash hashImpl
}

// WriteRlp writes an RLP value
func (k *Keccak) WriteRlp(dst []byte, v *fastrlp.Value) []byte {
	k.buf = v.MarshalTo(k.buf[:0])
	_, _ = k.Write(k.buf)
	return k.Sum(dst)
}

// Write implements the hash interface
func (k *Keccak) Write(b []byte) (int, error) {
	return k.hash.Write(b)
}

// Reset implements the hash interface
func (k *Keccak) Reset() {
	k.buf = k.buf[:0]
	k.hash.Reset()
}

// Read hashes the content and returns the intermediate buffer.
func (k *Keccak) Read() []byte {
	_, _ = k.hash.Read(k.tmp)
	return k.tmp
}

// Sum implements the hash interface
func (k *Keccak) Sum(dst []byte) []byte {
	_, _ = k.hash.Read(k.tmp)
	dst = append(dst, k.tmp[:]...)
	return dst
}

func newKeccak(hash hashImpl) *Keccak {
	return &Keccak{
		hash: hash,
		tmp:  make([]byte, hash.Size()),
	}
}

// NewKeccak256 returns a new keccak 256
func NewKeccak256() *Keccak {
	return newKeccak(sha3.NewLegacyKeccak256().(hashImpl))
}

// NewKeccak512 returns a new keccak 512
func NewKeccak512() *Keccak {
	return newKeccak(sha3.NewLegacyKeccak512().(hashImpl))
}
