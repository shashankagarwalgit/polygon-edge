package framework

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/command/rootchain/server"
)

type TestBridge struct {
	t             *testing.T
	clusterConfig *TestClusterConfig
	node          *node
}

func NewTestBridge(t *testing.T, clusterConfig *TestClusterConfig) (*TestBridge, error) {
	t.Helper()

	bridge := &TestBridge{
		t:             t,
		clusterConfig: clusterConfig,
	}

	err := bridge.Start()
	if err != nil {
		return nil, err
	}

	return bridge, nil
}

func (t *TestBridge) Start() error {
	// Build arguments
	args := []string{
		"rootchain",
		"server",
		"--data-dir", t.clusterConfig.Dir("test-rootchain"),
	}

	stdout := t.clusterConfig.GetStdout("bridge")

	bridgeNode, err := newNode(t.clusterConfig.Binary, args, stdout)
	if err != nil {
		return err
	}

	t.node = bridgeNode

	if err = server.PingServer(nil); err != nil {
		return err
	}

	return nil
}

func (t *TestBridge) Stop() {
	if err := t.node.Stop(); err != nil {
		t.t.Error(err)
	}

	t.node = nil
}

func (t *TestBridge) JSONRPCAddr() string {
	return fmt.Sprintf("http://%s:%d", hostIP, 8545)
}

func (t *TestBridge) WaitUntil(pollFrequency, timeout time.Duration, handler func() (bool, error)) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("timeout")
		case <-time.After(pollFrequency):
		}

		isConditionMet, err := handler()
		if err != nil {
			return err
		}

		if isConditionMet {
			return nil
		}
	}
}

// Deposit function is used to invoke bridge deposit command
// with appropriately created receivers and amounts for test transactions
func (t *TestBridge) Deposit(tokenType, receivers, amounts string) error {
	if tokenType == "" {
		return errors.New("provide token type value")
	}

	if receivers == "" {
		return errors.New("provide at least one receiver address value")
	}

	if amounts == "" {
		return errors.New("provide at least one amount value")
	}

	return t.cmdRun(
		"bridge",
		"deposit",
		"--manifest", path.Join(t.clusterConfig.TmpDir, "manifest.json"),
		"--token", tokenType,
		"--receivers", receivers,
		"--amounts", amounts)
}

// cmdRun executes arbitrary command from the given binary
func (t *TestBridge) cmdRun(args ...string) error {
	return runCommand(t.clusterConfig.Binary, args, t.clusterConfig.GetStdout("bridge"))
}

// deployRootchainContracts deploys and initializes rootchain contracts
func (t *TestBridge) deployRootchainContracts(manifestPath string) error {
	args := []string{
		"rootchain",
		"init-contracts",
		"--manifest", manifestPath,
	}

	if err := t.cmdRun(args...); err != nil {
		return fmt.Errorf("failed to deploy rootchain contracts: %w", err)
	}

	return nil
}

// fundRootchainValidators sends predefined amount of tokens to rootchain validators
func (t *TestBridge) fundRootchainValidators() error {
	args := []string{
		"rootchain",
		"fund",
		"--data-dir", path.Join(t.clusterConfig.TmpDir, t.clusterConfig.ValidatorPrefix),
		"--num", strconv.Itoa(int(t.clusterConfig.ValidatorSetSize) + t.clusterConfig.NonValidatorCount),
	}

	if err := t.cmdRun(args...); err != nil {
		return fmt.Errorf("failed to deploy fund validators: %w", err)
	}

	return nil
}
