package deployment

import (
	"fmt"
	"os"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/command"
	"github.com/0xPolygon/polygon-edge/command/helper"
	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/types"
)

const (
	chainFlag         = "chain"
	addAddressFlag    = "add-address"
	removeAddressFlag = "remove-address"
)

var (
	params = &deploymentParams{}
)

type deploymentParams struct {
	// raw addresses, entered by CLI commands
	addAddressRaw    []string
	removeAddressRaw []string

	// addresses, converted from raw addresses
	addAddress    []types.Address
	removeAddress []types.Address

	// genesis file
	genesisPath   string
	genesisConfig *chain.Chain

	// deployment whitelist from genesis configuration
	whitelist []types.Address
}

func (p *deploymentParams) initRawParams() error {
	// convert raw addresses to appropriate format
	if err := p.initRawAddresses(); err != nil {
		return err
	}

	// init genesis configuration
	if err := p.initChain(); err != nil {
		return err
	}

	return nil
}

func (p *deploymentParams) initRawAddresses() error {
	// convert addresses to be added from string to type.Address
	p.addAddress = unmarshallRawAddresses(p.addAddressRaw)

	// convert addresses to be removed from string to type.Address
	p.removeAddress = unmarshallRawAddresses(p.removeAddressRaw)

	return nil
}

func (p *deploymentParams) initChain() error {
	// import genesis configuration
	cc, err := chain.Import(p.genesisPath)
	if err != nil {
		return fmt.Errorf(
			"failed to load chain config from %s: %w",
			p.genesisPath,
			err,
		)
	}

	// set genesis configuration
	p.genesisConfig = cc

	return nil
}

func (p *deploymentParams) updateGenesisConfig() error {
	// Fetch contract deployment whitelist from genesis config
	deploymentWhitelist, err := common.FetchDeploymentWhitelist(p.genesisConfig)
	if err != nil {
		return err
	}

	// Add addresses if it doesn't exist
	for _, address := range p.addAddress {
		if !types.AddressExists(address, deploymentWhitelist) {
			deploymentWhitelist = append(deploymentWhitelist, address)
		}
	}

	newDeploymentWhitelist := make([]types.Address, 0)

	// Remove addresses if exists
	for _, address := range deploymentWhitelist {
		if !types.AddressExists(address, p.removeAddress) {
			newDeploymentWhitelist = append(newDeploymentWhitelist, address)
		}
	}

	// Set whitelist in genesis configuration
	whitelistConfig := common.FetchWhitelistFromConfig(p.genesisConfig)
	whitelistConfig[common.DeploymentWhitelistKey] = newDeploymentWhitelist
	p.genesisConfig.Params.Whitelists = whitelistConfig

	// Save whitelist for result
	p.whitelist = newDeploymentWhitelist

	return nil
}

func (p *deploymentParams) overrideGenesisConfig() error {
	// Remove the current genesis configuration from the disk
	if err := os.Remove(p.genesisPath); err != nil {
		return err
	}

	// Save the new genesis configuration
	if err := helper.WriteGenesisConfigToDisk(
		p.genesisConfig,
		p.genesisPath,
	); err != nil {
		return err
	}

	return nil
}

func (p *deploymentParams) getResult() command.CommandResult {
	result := &DeploymentResult{
		AddAddress:    p.addAddress,
		RemoveAddress: p.removeAddress,
		Whitelist:     p.whitelist,
	}

	return result
}

func unmarshallRawAddresses(addresses []string) []types.Address {
	marshalledAddresses := make([]types.Address, len(addresses))

	for indx, address := range addresses {
		marshalledAddresses[indx] = types.StringToAddress(address)
	}

	return marshalledAddresses
}
