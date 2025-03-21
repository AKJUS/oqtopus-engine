//go:build unit
// +build unit

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSettingPeople struct {
	PeopleNames []string `toml:"people_names"`
}

type TestSettingViecle struct {
	ViecleNames []string `toml:"viecle_names"`
}

func TestRegisterSettings(t *testing.T) {
	s := registeredSettings()
	assert.Equal(t, 2, len(s.ComponentSetting))
}

func TestParseSettings(t *testing.T) {
	ResetSetting()
	tests := []struct {
		name      string
		in        string
		wantError error
		want      *Setting
	}{
		{
			name:      "empty",
			in:        "",
			wantError: nil,
			want: &Setting{
				ComponentSetting: map[string]interface{}{},
				RunGroupSetting:  map[string]interface{}{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotError := globalSetting.parseSetting(tt.in)
			assert.Equal(t, tt.wantError, gotError)
			assert.Equal(t, tt.want, globalSetting)
		})
	}
}

func registeredSettings() *Setting {
	ns := newSetting()
	ns.registerSetting("people", &TestSettingPeople{
		PeopleNames: []string{},
	})
	ns.registerSetting("viecle", &TestSettingViecle{
		ViecleNames: []string{},
	})
	return ns
}
