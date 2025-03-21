//go:build unit
// +build unit

package qpu

import (
	"testing"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestDeviceSetting(t *testing.T) {
	blob, assetErr := common.GetAsset("unit_test_device_setting.toml")
	assert.Nil(t, assetErr)

	ds := DeviceSetting{}
	_, err := toml.Decode(blob, &ds)
	assert.Nil(t, err)
	assert.Equal(t, ds.DeviceName, "wako")

	assert.True(t, ds.QASMSupport.AllowList.Enabled)
	assert.False(t, ds.QASMSupport.DenyList.Enabled)

	allowStatements := ds.QASMSupport.AllowList.Statements
	assert.Contains(t, allowStatements, &QASMStatementType{Name: "if"})
	assert.Contains(t, allowStatements, &QASMStatementType{Name: "gate_call"})

	denyStatements := ds.QASMSupport.DenyList.Statements
	assert.Contains(t, denyStatements, &QASMStatementType{Name: "def"})
	assert.Contains(t, denyStatements, &QASMStatementType{Name: "gate_declaration"})
}
