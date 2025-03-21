//go:build unit
// +build unit

package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAsset(t *testing.T) {
	qasm, err := GetAsset("bell_pair.qasm")
	assert.Nil(t, err)
	assert.Equal(t, "OPENQASM 3;\nqubit[2] q;\nbit[2] c;\n\nh q[0];\ncx q[0], q[1];\n\nc[0] = measure q[0];\nc[1] = measure q[1];", qasm)
}

// TODO use TDT
func TestValidAddressWrongHost(t *testing.T) {
	host := "hogehoge^^^-server.com"
	port := "23413"
	address, err := ValidAddress(host, port)

	assert.EqualError(t, err, fmt.Sprintf("%s is an invalid host name", host))
	assert.Equal(t, address, "")
}
func TestValidAddressWrongPort(t *testing.T) {
	host := "hogehoge-server.com"
	port := "-23413"
	address, err := ValidAddress(host, port)

	assert.EqualError(t, err, fmt.Sprintf("%s is an invalid port number", port))
	assert.Equal(t, address, "")
}

func TestValidAddressWrongRangePort(t *testing.T) {
	host := "hogehoge-server.com"
	port := "23413431243214"
	address, err := ValidAddress(host, port)

	assert.EqualError(t, err, fmt.Sprintf("%s is not a port number within the allowed range", port))
	assert.Equal(t, address, "")
}

func TestPlaninJsonString(t *testing.T) {
	jsonString := "{\n  \"name\": \"wako\",\n  \"qubits\"}"
	expected := "{\"name\":\"wako\",\"qubits\"}"

	actual := PlainJsonString(jsonString)
	assert.Equal(t, expected, actual)
}
