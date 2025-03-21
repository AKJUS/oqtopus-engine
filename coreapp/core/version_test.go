//go:build unit
// +build unit

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name               string
		conf               *Conf
		versionByBuildFlag string
		wantVersion        string
	}{
		{
			name:               "version is set by build flag",
			conf:               &Conf{},
			versionByBuildFlag: "v1.2.3",
			wantVersion:        "v1.2.3",
		},
		{
			name:               "version is set by config",
			conf:               &Conf{Version: "v1.2.3"},
			versionByBuildFlag: "",
			wantVersion:        "v1.2.3",
		},
		{
			name:               "version is not set",
			conf:               &Conf{},
			versionByBuildFlag: "",
			wantVersion:        NoVersion,
		},
		{
			name:               "version is set by build flag and config",
			conf:               &Conf{Version: "v1.2.3"},
			versionByBuildFlag: "v1.2.4",
			wantVersion:        "v1.2.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetVersion(tt.conf, tt.versionByBuildFlag)
			assert.Equal(t, Version, tt.wantVersion)
		})
	}
}
