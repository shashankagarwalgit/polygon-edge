package loadbot

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/0xPolygon/polygon-sdk/crypto"
	txPoolOp "github.com/0xPolygon/polygon-sdk/txpool/proto"
	"github.com/0xPolygon/polygon-sdk/types"
	"github.com/umbracle/go-web3/jsonrpc"
	"google.golang.org/grpc"
	"math/big"
	"sync"
	"time"
)

type Configuration struct {
	TPS           uint64
	AccountsCount uint64
	Value         int64
	Count         uint64
	JSONRPCs      []string
	GRPCs         []string
}

type Metrics struct {
	m                          sync.Mutex
	Duration                   time.Duration
	TotalTransactionsSentCount uint64
	FailedTransactionsCount    uint64
}

type Account struct {
	Address    types.Address
	PrivateKey ecdsa.PrivateKey
}

// generateRandomValue creates a random value used in a transaction.
// The max value that can be generated represents 0.01 ETH.
func generateRandomValue() (*big.Int, error) {
	b, err := rand.Int(rand.Reader, big.NewInt(10000000000000000))
	if err != nil {
		return nil, fmt.Errorf("failed to create random number: %v", err)
	}
	return b, nil
}

func createJsonRpcClient(endpoint string) (*jsonrpc.Client, error) {
	client, err := jsonrpc.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create new JSON RPC client: %v", err)
	}
	return client, nil
}

func createGRpcClient(endpoint string) (*txPoolOp.TxnPoolOperatorClient, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %v", err)
	}

	client := txPoolOp.NewTxnPoolOperatorClient(conn)
	return &client, nil
}

func generateAccounts(n uint64) ([]*Account, error) {
	var accounts []*Account

	for i := uint64(0); i < n; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create ecdsa key pair: %v", err)
		}

		account := Account{
			Address:    crypto.PubKeyToAddress(&privateKey.PublicKey),
			PrivateKey: *privateKey,
		}

		accounts = append(accounts, &account)
	}
	return accounts, nil
}

func execute() error {
	// Get sender and receiver accounts
	// Get nonce for the sender account
	// If required, generate new value for the transaction
	// Create the transaction object
	return nil
}

func Run(conf *Configuration) (error, *Metrics) {
	// Create the ticker
	ticker := time.NewTicker(1 * time.Second / time.Duration(conf.TPS))
	defer ticker.Stop()

	// Create the metrics placeholder
	metrics := Metrics{
		Duration:                   0,
		TotalTransactionsSentCount: 0,
		FailedTransactionsCount:    0,
	}

	// Record execution time
	start := time.Now()
	defer func() {
		metrics.Duration = time.Since(start)
	}()

	// Loop and send a transaction at each tick
	for {
		select {
		case <-ticker.C:
			// Register new operation in the metrics
			metrics.m.Lock()
			metrics.TotalTransactionsSentCount += 1
			metrics.m.Unlock()

			err := execute()

			// Register an error in the metrics
			if err != nil {
				metrics.m.Lock()
				metrics.FailedTransactionsCount += 1
				metrics.m.Unlock()
			}
			return nil, nil
		}
	}
}
