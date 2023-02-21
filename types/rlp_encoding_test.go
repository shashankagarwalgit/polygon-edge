package types

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/umbracle/fastrlp"

	"github.com/stretchr/testify/assert"
)

type codec interface {
	RLPMarshaler
	RLPUnmarshaler
}

func TestRLPEncoding(t *testing.T) {
	cases := []codec{
		&Header{},
		&Receipt{},
	}
	for _, c := range cases {
		buf := c.MarshalRLPTo(nil)

		res, ok := reflect.New(reflect.TypeOf(c).Elem()).Interface().(codec)
		if !ok {
			t.Fatalf("Unable to assert type")
		}

		if err := res.UnmarshalRLP(buf); err != nil {
			t.Fatal(err)
		}

		buf2 := c.MarshalRLPTo(nil)
		if !reflect.DeepEqual(buf, buf2) {
			t.Fatal("[ERROR] Buffers not equal")
		}
	}
}

func TestRLPMarshall_And_Unmarshall_Transaction(t *testing.T) {
	addrTo := StringToAddress("11")
	txn := &Transaction{
		Nonce:    0,
		GasPrice: big.NewInt(11),
		Gas:      11,
		To:       &addrTo,
		Value:    big.NewInt(1),
		Input:    []byte{1, 2},
		V:        big.NewInt(25),
		S:        big.NewInt(26),
		R:        big.NewInt(27),
	}
	unmarshalledTxn := new(Transaction)
	marshaledRlp := txn.MarshalRLP()

	if err := unmarshalledTxn.UnmarshalRLP(marshaledRlp); err != nil {
		t.Fatal(err)
	}

	unmarshalledTxn.ComputeHash()

	txn.Hash = unmarshalledTxn.Hash
	assert.Equal(t, txn, unmarshalledTxn, "[ERROR] Unmarshalled transaction not equal to base transaction")
}

func TestRLPStorage_Marshall_And_Unmarshall_Receipt(t *testing.T) {
	addr := StringToAddress("11")
	hash := StringToHash("10")

	testTable := []struct {
		name      string
		receipt   *Receipt
		setStatus bool
	}{
		{
			"Marshal receipt with status",
			&Receipt{
				CumulativeGasUsed: 10,
				GasUsed:           100,
				ContractAddress:   &addr,
				TxHash:            hash,
			},
			true,
		},
		{
			"Marshal receipt without status",
			&Receipt{
				Root:              hash,
				CumulativeGasUsed: 10,
				GasUsed:           100,
				ContractAddress:   &addr,
				TxHash:            hash,
			},
			false,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			receipt := testCase.receipt

			if testCase.setStatus {
				receipt.SetStatus(ReceiptSuccess)
			}

			unmarshalledReceipt := new(Receipt)
			marshaledRlp := receipt.MarshalStoreRLPTo(nil)

			if err := unmarshalledReceipt.UnmarshalStoreRLP(marshaledRlp); err != nil {
				t.Fatal(err)
			}

			if !assert.Exactly(t, receipt, unmarshalledReceipt) {
				t.Fatal("[ERROR] Unmarshalled receipt not equal to base receipt")
			}
		})
	}
}

func TestRLPUnmarshal_Header_ComputeHash(t *testing.T) {
	// header computes hash after unmarshalling
	h := &Header{}
	h.ComputeHash()

	data := h.MarshalRLP()
	h2 := new(Header)
	assert.NoError(t, h2.UnmarshalRLP(data))
	assert.Equal(t, h.Hash, h2.Hash)
}

func TestRLPMarshall_And_Unmarshall_TypedTransaction(t *testing.T) {
	addrTo := StringToAddress("11")
	addrFrom := StringToAddress("22")
	originalTx := &Transaction{
		Nonce:    0,
		GasPrice: big.NewInt(11),
		Gas:      11,
		To:       &addrTo,
		From:     addrFrom,
		Value:    big.NewInt(1),
		Input:    []byte{1, 2},
		V:        big.NewInt(25),
		S:        big.NewInt(26),
		R:        big.NewInt(27),
	}

	txTypes := []TxType{
		StateTx,
		LegacyTx,
	}

	for _, v := range txTypes {
		originalTx.Type = v
		originalTx.ComputeHash()

		txRLP := originalTx.MarshalRLP()

		unmarshalledTx := new(Transaction)
		assert.NoError(t, unmarshalledTx.UnmarshalRLP(txRLP))

		unmarshalledTx.ComputeHash()
		assert.Equal(t, originalTx.Type, unmarshalledTx.Type)
	}
}

func TestRLPMarshall_And_Unmarshall_TxType(t *testing.T) {
	testTable := []struct {
		name        string
		txType      TxType
		expectedErr bool
	}{
		{
			name:   "StateTx",
			txType: StateTx,
		},
		{
			name:   "LegacyTx",
			txType: LegacyTx,
		},
		{
			name:        "undefined type",
			txType:      TxType(0x09),
			expectedErr: true,
		},
	}

	for _, tt := range testTable {
		ar := &fastrlp.Arena{}

		var txType TxType
		err := txType.unmarshalRLPFrom(nil, ar.NewBytes([]byte{byte(tt.txType)}))

		if tt.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.txType, txType)
		}
	}
}
