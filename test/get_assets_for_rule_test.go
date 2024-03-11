package test

import (
	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAssetsForRule(t *testing.T) {
	config := &model.Config{
		Wallets: []model.Wallet{
			{
				Name:     "ETH_cold",
				Asset:    "ETH",
				WalletId: "wallet123",
			},
			{
				Name:     "BTC_hot",
				Asset:    "BTC",
				WalletId: "wallet456",
			},
		},
	}

	tests := []struct {
		name     string
		rule     model.Rule
		expected []string
	}{
		{
			name: "single wallet",
			rule: model.Rule{
				Wallets: []string{"ETH_cold"},
			},
			expected: []string{"ETH"},
		},
		{
			name: "multiple wallets",
			rule: model.Rule{
				Wallets: []string{"ETH_cold", "BTC_hot"},
			},
			expected: []string{"ETH", "BTC"},
		},
		{
			name: "no matching wallets",
			rule: model.Rule{
				Wallets: []string{"XRP_cold"},
			},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := core.GetAssetsForRule(tc.rule, config)
			if len(tc.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}
