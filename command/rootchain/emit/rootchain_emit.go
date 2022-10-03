package emit

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/jsonrpc"
	"golang.org/x/sync/errgroup"

	"github.com/0xPolygon/polygon-edge/command"
	"github.com/0xPolygon/polygon-edge/command/rootchain/helper"
	smartcontracts "github.com/0xPolygon/polygon-edge/contracts/smart_contracts"
	"github.com/0xPolygon/polygon-edge/types"
)

var (
	params emitParams

	contractsToParamTypes = map[string]string{
		helper.SidechainBridgeAddr.String(): "tuple(address,uint256)",
	}
)

// GetCommand returns the rootchain emit command
func GetCommand() *cobra.Command {
	rootchainEmitCmd := &cobra.Command{
		Use:     "emit",
		Short:   "Emit an event from the bridge",
		PreRunE: runPreRun,
		Run:     runCommand,
	}

	setFlags(rootchainEmitCmd)

	return rootchainEmitCmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&params.contractAddrRaw,
		contractFlag,
		helper.SidechainBridgeAddr.String(),
		"ERC20 bridge contract address",
	)

	cmd.Flags().StringSliceVar(
		&params.wallets,
		walletsFlag,
		nil,
		"list of wallet addresses",
	)

	cmd.Flags().StringSliceVar(
		&params.amounts,
		amountsFlag,
		nil,
		"list of amounts to fund wallets",
	)
}

func runPreRun(_ *cobra.Command, _ []string) error {
	return params.validateFlags()
}

func runCommand(cmd *cobra.Command, _ []string) {
	outputter := command.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	paramsType, exists := contractsToParamTypes[params.contractAddrRaw]
	if !exists {
		outputter.SetError(fmt.Errorf("there are no parameter types registered for given contract address: %v", params.contractAddrRaw))
		return
	}

	ipAddr := helper.ReadRootchainIP()
	rpcClient, err := jsonrpc.NewClient(ipAddr)
	if err != nil {
		outputter.SetError(fmt.Errorf("could not establish new json rpc client: %s", err))
		return
	}

	pendingNonce, err := rpcClient.Eth().GetNonce(helper.GetDefAccount(), ethgo.Pending)
	if err != nil {
		outputter.SetError(fmt.Errorf("could not get pending nonce: %s", err))
		return
	}

	g, ctx := errgroup.WithContext(context.Background())
	for i := range params.wallets {
		i := i // goroutine closure
		wallet := params.wallets[i]
		amount := params.amounts[i]
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				nonce := pendingNonce + uint64(i)
				txn := createTxInput(paramsType, wallet, amount)
				if _, err = helper.SendTxn(rpcClient, nonce, txn); err != nil {
					return fmt.Errorf("sending transaction to wallet: %s with amount: %s, failed with error: %w", wallet, amount, err)
				}
				return nil
			}
		})
	}

	if err = g.Wait(); err != nil {
		outputter.SetError(fmt.Errorf("sending transactions to rootchain failed: %s", err))
		return
	}

	outputter.SetCommandResult(&RootchainEmitResult{
		ContractAddr: params.contractAddrRaw,
		Wallets:      params.wallets,
		Amounts:      params.amounts,
	})
}

func createTxInput(paramsType string, parameters ...interface{}) *ethgo.Transaction {
	var prms []interface{}
	prms = append(prms, parameters...)

	wrapperInput, err := abi.MustNewType(paramsType).Encode(prms)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode parsed parameters. Error: %v", err))
	}

	artifact := smartcontracts.MustReadArtifact("rootchain", "RootchainBridge")
	method := artifact.Abi.Methods["emitEvent"]
	input, err := method.Encode([]interface{}{types.StringToAddress(params.contractAddrRaw), wrapperInput})
	if err != nil {
		panic(fmt.Sprintf("Failed to encode provided parameters. Error: %v", err))
	}

	return &ethgo.Transaction{
		To:    (*ethgo.Address)(&helper.RootchainBridgeAddress),
		Input: input,
	}
}
