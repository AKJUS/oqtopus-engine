package qpu

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"go.uber.org/zap"
)

// TODO: remove DeviceSetting, instead use DefaultGatewayAgentSetting
type DeviceSetting struct {
	DeviceName    string       `toml:"device_name"`
	DeviceType    string       `toml:"device_type"`
	ProviderName  string       `toml:"provider_name"`
	MaxShots      int          `toml:"max_shots"`
	QASMSupport   *QASMSupport `toml:"qasm_support"`
	MachineHost   string       `toml:"machine_host"`
	MachinePort   string       `toml:"machine_port"`
	PollingPeriod uint32       `toml:"polling_period"`
	UseCred       bool         `toml:"use_cred"`
}

type QASMSupport struct {
	AllowList *QASMFilter `toml:"allow_list"`
	DenyList  *QASMFilter `toml:"deny_list"`
}

func LoadDeviceSetting(path string) (*DeviceSetting, error) {
	blob, assetErr := common.ReadFile(path)
	ds := NewDeviceSetting()
	if assetErr != nil {
		zap.L().Info(fmt.Sprintf("Failed to read file:%s Reason:%s", path, assetErr))
		return ds, nil
	}
	if _, err := toml.Decode(blob, ds); err != nil {
		zap.L().Error(fmt.Sprintf("failed to decode blob:%s", blob))
		return &DeviceSetting{}, err
	}
	return ds, nil
}

func NewDeviceSetting() *DeviceSetting {
	return &DeviceSetting{
		QASMSupport:   NewQasmSupport(),
		MachineHost:   "localhost",
		MachinePort:   "50051",
		PollingPeriod: 60,
	}
}

func NewQasmSupport() *QASMSupport {
	return &QASMSupport{
		AllowList: &QASMFilter{},
		DenyList:  &QASMFilter{},
	}
}

func NewQasmSupportWithAllowList(q *QASMFilter) *QASMSupport {
	return &QASMSupport{
		AllowList: q,
		DenyList:  &QASMFilter{},
	}
}

func NewQasmSupportWithDenyList(q *QASMFilter) *QASMSupport {
	return &QASMSupport{
		AllowList: &QASMFilter{},
		DenyList:  q,
	}
}

type QASMFilter struct {
	Enabled    bool
	Statements []*QASMStatementType `toml:"statements"`
	Gates      []*QASMGateType      `toml:"gates"`
}
