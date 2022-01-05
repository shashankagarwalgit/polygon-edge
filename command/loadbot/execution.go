package loadbot

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/0xPolygon/polygon-sdk/command/loadbot/generator"
	txpoolOp "github.com/0xPolygon/polygon-sdk/txpool/proto"
	"github.com/golang/protobuf/ptypes/any"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/polygon-sdk/helper/tests"
	"github.com/0xPolygon/polygon-sdk/types"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

const (
	maxReceiptWait = 5 * time.Minute
	minReceiptWait = 30 * time.Second

	defaultFastestTurnAround = time.Hour * 24
	defaultSlowestTurnAround = time.Duration(0)
)

type Account struct {
	Address    types.Address
	PrivateKey *ecdsa.PrivateKey
}
type Configuration struct {
	TPS           uint64
	Sender        types.Address
	Receiver      types.Address
	Value         *big.Int
	Count         uint64
	JSONRPC       string
	GRPC          string
	MaxConns      int
	GeneratorMode uint64
	ChainID       uint64
}

type metadata struct {
	// turn around time for the transaction
	turnAroundTime time.Duration

	// block where it was sealed
	blockNumber uint64
}

type ExecDuration struct {
	// turnAroundMap maps the transaction hash -> turn around time for passing transactions
	turnAroundMap     sync.Map
	turnAroundMapSize uint64

	// blockTransactions maps how many transactions went into a block
	blockTransactions map[uint64]uint64

	// Arrival Time - Time at which the transaction is added
	// Completion Time -Time at which the transaction is sealed
	// Turn around time - Completion Time – Arrival Time

	// AverageTurnAround is the average turn around time for all passing transactions
	AverageTurnAround time.Duration

	// FastestTurnAround is the fastest turn around time recorded for a transaction
	FastestTurnAround time.Duration

	// SlowestTurnAround is the slowest turn around time recorded for a transaction
	SlowestTurnAround time.Duration

	// TotalExecTime is the total execution time for a single loadbot run
	TotalExecTime time.Duration
}

// calcTurnAroundMetrics updates the turn around metrics based on the turnAroundMap
func (ed *ExecDuration) calcTurnAroundMetrics() {
	// Set the initial values
	fastestTurnAround := defaultFastestTurnAround
	slowestTurnAround := defaultSlowestTurnAround
	totalPassing := atomic.LoadUint64(&ed.turnAroundMapSize)
	var zeroTime time.Time  // Zero time
	var totalTime time.Time // Zero time used for tracking

	if totalPassing == 0 {
		// No data to show, use zero data
		zeroDuration := time.Duration(0)
		ed.SlowestTurnAround = zeroDuration
		ed.FastestTurnAround = zeroDuration
		ed.AverageTurnAround = zeroDuration

		return
	}

	ed.turnAroundMap.Range(func(_, value interface{}) bool {
		data := value.(*metadata)
		turnAroundTime := data.turnAroundTime

		// Update the duration metrics
		if turnAroundTime < fastestTurnAround {
			fastestTurnAround = turnAroundTime
		}

		if turnAroundTime > slowestTurnAround {
			slowestTurnAround = turnAroundTime
		}

		totalTime = totalTime.Add(turnAroundTime)

		ed.blockTransactions[data.blockNumber]++

		return true
	})

	averageDuration := (totalTime.Sub(zeroTime)) / time.Duration(totalPassing)

	ed.SlowestTurnAround = slowestTurnAround
	ed.FastestTurnAround = fastestTurnAround
	ed.AverageTurnAround = averageDuration
}

// reportExecTime reports the turn around time for a transaction
// for a single loadbot run
func (ed *ExecDuration) reportTurnAroundTime(
	txHash web3.Hash,
	data *metadata,
) {
	ed.turnAroundMap.Store(txHash, data)
	atomic.AddUint64(&ed.turnAroundMapSize, 1)
}

type Metrics struct {
	TotalTransactionsSentCount uint64
	FailedTransactionsCount    uint64
	TransactionDuration        ExecDuration
}

type Loadbot struct {
	cfg       *Configuration
	metrics   *Metrics
	generator generator.TransactionGenerator
}

// calcMaxTimeout calculates the max timeout for transactions receipts
// based on the transaction count and tps params
func calcMaxTimeout(count, tps uint64) time.Duration {
	waitTime := minReceiptWait
	// The receipt timeout should be at max maxReceiptWait
	// or minReceiptWait + tps / count * 100
	// This way the wait time scales linearly for more stressful situations
	waitFactor := time.Duration(float64(tps)/float64(count)*100) * time.Second

	if waitTime+waitFactor > maxReceiptWait {
		return maxReceiptWait
	}

	return waitTime + waitFactor
}

func NewLoadBot(cfg *Configuration, metrics *Metrics) *Loadbot {
	return &Loadbot{
		cfg:     cfg,
		metrics: metrics,
	}
}

func getInitialSenderNonce(client *jsonrpc.Client, address types.Address) (uint64, error) {
	nonce, err := client.Eth().GetNonce(web3.Address(address), web3.Latest)
	if err != nil {
		return 0, fmt.Errorf("failed to query initial sender nonce: %v", err)
	}
	return nonce, nil
}

func (l *Loadbot) executeTxn(
	client txpoolOp.TxnPoolOperatorClient,
) (web3.Hash, error) {
	txn, err := l.generator.GenerateTransaction()
	if err != nil {
		return web3.Hash{}, err
	}

	addReq := &txpoolOp.AddTxnReq{
		Raw: &any.Any{
			Value: txn.MarshalRLP(),
		},
		From: types.ZeroAddress.String(),
	}

	addRes, addErr := client.AddTxn(context.Background(), addReq)
	if addErr != nil {
		return web3.Hash{}, fmt.Errorf("unable to add transaction, %v", addErr)
	}

	return web3.Hash(types.StringToHash(addRes.TxHash)), nil
}

func (l *Loadbot) Run() error {
	sender, err := extractSenderAccount(l.cfg.Sender)
	if err != nil {
		return fmt.Errorf("failed to extract sender account: %v", err)
	}

	jsonClient, err := createJsonRpcClient(l.cfg.JSONRPC, l.cfg.MaxConns)
	if err != nil {
		return fmt.Errorf("an error has occured while creating JSON-RPC client: %v", err)
	}

	grpcClient, err := createGRPCClient(l.cfg.GRPC)
	if err != nil {
		return fmt.Errorf("an error has occured while creating JSON-RPC client: %v", err)
	}

	defer func(client *jsonrpc.Client) {
		_ = shutdownClient(client)
	}(jsonClient)

	nonce, err := getInitialSenderNonce(jsonClient, sender.Address)
	if err != nil {
		return fmt.Errorf("an error occured while getting initial sender nonce: %v", err)
	}

	// Set up the transaction generator
	generatorParams := &generator.GeneratorParams{
		Nonce:         nonce,
		ChainID:       l.cfg.ChainID,
		SenderAddress: sender.Address,
		SenderKey:     sender.PrivateKey,
		Value:         l.cfg.Value,
		EstimatedGas:  1000000, // TODO add gas estimation
	}

	var txnGenerator generator.TransactionGenerator
	var genErr error = nil
	switch l.cfg.GeneratorMode {
	case 0:
		txnGenerator, genErr = generator.NewTransferGenerator(generatorParams)
	case 1:
		txnGenerator, genErr = generator.NewDeployGenerator(generatorParams)
	}

	if genErr != nil {
		return fmt.Errorf("unable to start generator, %v", genErr)
	}
	l.generator = txnGenerator

	ticker := time.NewTicker(1 * time.Second / time.Duration(l.cfg.TPS))
	defer ticker.Stop()

	var wg sync.WaitGroup

	receiptTimeout := calcMaxTimeout(l.cfg.Count, l.cfg.TPS)

	startTime := time.Now()
	for i := uint64(0); i < l.cfg.Count; i++ {
		<-ticker.C

		l.metrics.TotalTransactionsSentCount += 1

		wg.Add(1)
		go func(index uint64) {
			defer wg.Done()

			// Start the performance timer
			start := time.Now()

			// Execute the transaction
			txHash, err := l.executeTxn(grpcClient)
			if err != nil {
				l.generator.MarkFailedTxn(&generator.FailedTxnInfo{
					Index:  index,
					TxHash: txHash.String(),
					Error: &generator.TxnError{
						Error:     err,
						ErrorType: generator.AddErrorType,
					},
				})
				atomic.AddUint64(&l.metrics.FailedTransactionsCount, 1)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), receiptTimeout)
			defer cancel()

			receipt, err := tests.WaitForReceipt(ctx, jsonClient.Eth(), txHash)
			if err != nil {
				l.generator.MarkFailedTxn(&generator.FailedTxnInfo{
					Index:  index,
					TxHash: txHash.String(),
					Error: &generator.TxnError{
						Error:     err,
						ErrorType: generator.ReceiptErrorType,
					},
				})
				atomic.AddUint64(&l.metrics.FailedTransactionsCount, 1)
				return
			}

			// Stop the performance timer
			end := time.Now()
			l.metrics.TransactionDuration.reportTurnAroundTime(
				txHash,
				&metadata{
					turnAroundTime: end.Sub(start),
					blockNumber:    receipt.BlockNumber,
				},
			)
		}(i)
	}

	wg.Wait()
	endTime := time.Now()

	// Calculate the turn around metrics now that the loadbot is done
	l.metrics.TransactionDuration.calcTurnAroundMetrics()
	l.metrics.TransactionDuration.TotalExecTime = endTime.Sub(startTime)

	return nil
}
