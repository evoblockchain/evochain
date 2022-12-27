package types

import (
	"strings"
	"testing"

	"github.com/evoblockchain/evochain/libs/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// test compate 0x address to send and query
func TestAccAddressFromBech32(t *testing.T) {
	config := types.GetConfig()
	SetBech32Prefixes(config)

	//make data
	tests := []struct {
		addrStr    string
		expectPass bool
	}{
		{"evo1ss60jyagpw8v224006cqdkw22m32ytz54av3h3", true},
		{"0x8434f913a80B8EC52AAF7eb006d9Ca56e2a22c54", true},
		{"8434f913a80B8EC52AAF7eb006d9Ca56e2a22c54", true},
		{strings.ToLower("8434f913a80B8EC52AAF7eb006d9Ca56e2a22c54"), true},
		{strings.ToUpper("8434f913a80B8EC52AAF7eb006d9Ca56e2a22c54"), true},
		{"evo16zra6y26jytss520xrffq96p4fp4mnantcvyp2_", false},
		{"0xd087Dd115a911708514F30d2901741Aa435dCfB3_", false},
		{"d087Dd115a911708514F30d2901741Aa435dCfB3_", false},
	}

	//test run
	for _, tc := range tests {
		addr, err := types.AccAddressFromBech32(tc.addrStr)
		if tc.expectPass {
			require.NotNil(t, addr, "test: %v", tc.addrStr)
			require.Nil(t, err, "test: %v", tc.addrStr)
		} else {
			require.Nil(t, addr, "test: %v", tc.addrStr)
			require.NotNil(t, err, "test: %v", tc.addrStr)
		}
	}

}
