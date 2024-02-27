package test

import (
	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterWalletsByName(t *testing.T) {
	testConfig := &model.Config{
		Wallets: []model.Wallet{
			{
				Name:     "WalletA",
				WalletId: "d4cdc9a7-8901-446e-aebc-25a2715f9b74",
			},
			{
				Name:     "WalletB",
				WalletId: "id2",
			},
			{
				Name:     "WalletC",
				WalletId: "id3",
			},
		},
	}

	tests := []struct {
		name              string
		walletNames       []string
		expectedWalletIds []string
	}{
		{
			name:              "Filter single wallet",
			walletNames:       []string{"WalletA"},
			expectedWalletIds: []string{"d4cdc9a7-8901-446e-aebc-25a2715f9b74"},
		},
		{
			name:              "Filter multiple wallets",
			walletNames:       []string{"WalletA", "WalletC"},
			expectedWalletIds: []string{"d4cdc9a7-8901-446e-aebc-25a2715f9b74", "id3"},
		},
		{
			name:              "Filter non-existing wallet",
			walletNames:       []string{"WalletX"},
			expectedWalletIds: []string{},
		},
		{
			name:              "Empty filter list",
			walletNames:       []string{},
			expectedWalletIds: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filteredWalletIds := core.FilterWalletsByName(tc.walletNames, testConfig)
			if len(tc.expectedWalletIds) == 0 && len(filteredWalletIds) == 0 {
				assert.True(t, true, "Both expected and actual slices are empty")
			} else {
				assert.Equal(t, tc.expectedWalletIds, filteredWalletIds, "Filtered wallet IDs should match expected output")
			}
		})
	}
}
