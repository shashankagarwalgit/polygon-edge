package unstaking

import (
	"bytes"
	"fmt"

	"github.com/0xPolygon/polygon-edge/command/helper"
	sidechainHelper "github.com/0xPolygon/polygon-edge/command/sidechain"
)

var (
	undelegateAddressFlag = "undelegate"
)

type unstakeParams struct {
	accountDir    string
	accountConfig string
	jsonRPC       string
	amount        uint64
}

func (v *unstakeParams) validateFlags() error {
	return sidechainHelper.ValidateSecretFlags(v.accountDir, v.accountConfig)
}

type unstakeResult struct {
	validatorAddress string
	amount           uint64
}

func (ur unstakeResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[UNSTAKE]\n")

	vals := make([]string, 0, 2)
	vals = append(vals, fmt.Sprintf("Validator Address|%s", ur.validatorAddress))
	vals = append(vals, fmt.Sprintf("Amount Unstaked|%v", ur.amount))

	buffer.WriteString(helper.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}
