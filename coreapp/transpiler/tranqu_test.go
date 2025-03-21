//go:build unit
// +build unit

package transpiler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToPhysicalVirtualMappingFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[uint32]uint32
		wantErr bool
	}{
		{
			name:    "Success",
			input:   `{"qubit_mapping": {"0": 1, "1": 0}, "bit_mapping": {}}`,
			want:    map[uint32]uint32{1: 0, 0: 1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toPhysicalVirtualMappingFromString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			for k, v := range tt.want {
				assert.Equal(t, v, got[k])
			}
		})
	}
}
