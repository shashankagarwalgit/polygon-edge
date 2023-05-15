package tests

import (
	"bytes"
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

const (
	stateTests       = "tests/GeneralStateTests"
	legacyStateTests = "tests/LegacyTests/Constantinople/GeneralStateTests"
)

var (
	ripemd = types.StringToAddress("0000000000000000000000000000000000000003")
)

type stateCase struct {
	Env         *env                                    `json:"env"`
	Pre         map[types.Address]*chain.GenesisAccount `json:"pre"`
	Post        map[string]postState                    `json:"post"`
	Transaction *stTransaction                          `json:"transaction"`
}

func RunSpecificTest(t *testing.T, file string, c stateCase, name, fork string, index int, p postEntry) {
	t.Helper()

	config, ok := Forks[fork]
	if !ok {
		t.Skipf("%s fork is not supported", fork)

		return
	}

	env := c.Env.ToEnv(t)

	var baseFee *big.Int
	if config.IsLondon(0) {
		if c.Env.BaseFee != "" {
			baseFee = stringToBigIntT(t, c.Env.BaseFee)
		} else {
			// Retesteth uses `0x10` for genesis baseFee. Therefore, it defaults to
			// parent - 2 : 0xa as the basefee for 'this' context.
			baseFee = big.NewInt(0x0a)
		}
	}

	msg, err := c.Transaction.At(p.Indexes, baseFee)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	s, snapshot, pastRoot := buildState(c.Pre)
	forks := config.At(uint64(env.Number))

	xxx := state.NewExecutor(&chain.Params{
		Forks:   config,
		ChainID: 1,
		BurnContract: map[uint64]string{
			0: types.ZeroAddress.String(),
		},
	}, s, hclog.NewNullLogger())

	xxx.PostHook = func(t *state.Transition) {
		if name == "failed_tx_xcf416c53" {
			// create the account
			t.Txn().TouchAccount(ripemd)
			// now remove it
			t.Txn().Suicide(ripemd)
		}
	}
	xxx.GetHash = func(*types.Header) func(i uint64) types.Hash {
		return vmTestBlockHash
	}

	executor, _ := xxx.BeginTxn(pastRoot, c.Env.ToHeader(t), env.Coinbase)
	executor.Apply(msg) //nolint:errcheck

	txn := executor.Txn()

	// mining rewards
	txn.AddSealingReward(env.Coinbase, big.NewInt(0))

	objs := txn.Commit(forks.EIP155)
	_, root := snapshot.Commit(objs)

	if !bytes.Equal(root, p.Root.Bytes()) {
		t.Fatalf(
			"root mismatch (%s %s %s %d): expected %s but found %s",
			file,
			name,
			fork,
			index,
			p.Root.String(),
			hex.EncodeToHex(root),
		)
	}

	if logs := rlpHashLogs(txn.Logs()); logs != p.Logs {
		t.Fatalf(
			"logs mismatch (%s, %s %d): expected %s but found %s",
			name,
			fork,
			index,
			p.Logs.String(),
			logs.String(),
		)
	}
}

func TestState(t *testing.T) {
	t.Parallel()

	long := []string{
		"static_Call50000",
		"static_Return50000",
		"static_Call1MB",
		"stQuadraticComplexityTest",
		"stTimeConsuming",
	}

	skip := []string{
		"RevertPrecompiledTouch",
	}

	// There are two folders in spec tests, one for the current tests for the Istanbul fork
	// and one for the legacy tests for the other forks
	folders, err := listFolders(stateTests, legacyStateTests)
	if err != nil {
		t.Fatal(err)
	}

	for _, folder := range folders {
		folder := folder
		t.Run(folder, func(t *testing.T) {
			t.Parallel()

			files, err := listFiles(folder)
			if err != nil {
				t.Fatal(err)
			}

			for _, file := range files {
				if !strings.HasSuffix(file, ".json") {
					continue
				}

				if contains(long, file) && testing.Short() {
					t.Skipf("Long tests are skipped in short mode")

					continue
				}

				if contains(skip, file) {
					t.Skip()

					continue
				}

				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}

				var c map[string]stateCase
				if err = json.Unmarshal(data, &c); err != nil {
					t.Fatal(err)
				}

				for name, i := range c {
					for fork, f := range i.Post {
						for indx, e := range f {
							RunSpecificTest(t, file, i, name, fork, indx, e)
						}
					}
				}
			}
		})
	}
}
