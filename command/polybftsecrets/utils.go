package polybftsecrets

import (
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/secrets/helper"
)

// common flags for all polybft commands
const (
	DataPathFlag = "data-dir"
	ConfigFlag   = "config"

	DataPathFlagDesc = "the directory for the Polygon Edge data if the local FS is used"
	ConfigFlagDesc   = "the path to the SecretsManager config file, if omitted, the local FS secrets manager is used"
)

// common errors for all polybft commands
var (
	ErrInvalidNum                     = fmt.Errorf("num flag value should be between 1 and %d", maxInitNum)
	ErrInvalidConfig                  = errors.New("invalid secrets configuration")
	ErrInvalidParams                  = errors.New("no config file or data directory passed in")
	ErrUnsupportedType                = errors.New("unsupported secrets manager")
	ErrSecureLocalStoreNotImplemented = errors.New(
		"use a secrets backend, or supply an --insecure flag " +
			"to store the private keys locally on the filesystem, " +
			"avoid doing so in production")
)

func GetSecretsManager(dataPath, configPath string, insecureLocalStore bool) (secrets.SecretsManager, error) {
	if configPath != "" {
		secretsConfig, readErr := secrets.ReadConfig(configPath)
		if readErr != nil {
			return nil, ErrInvalidConfig
		}

		if !secrets.SupportedServiceManager(secretsConfig.Type) {
			return nil, ErrUnsupportedType
		}

		return helper.InitCloudSecretsManager(secretsConfig)
	}

	//Storing secrets on a local file system should only be allowed with --insecure flag,
	//to raise awareness that it should be only used in development/testing environments.
	//Production setups should use one of the supported secrets managers
	if !insecureLocalStore {
		return nil, ErrSecureLocalStoreNotImplemented
	}

	return helper.SetupLocalSecretsManager(dataPath)
}
