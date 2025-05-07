/**
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"testing"

	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/stretchr/testify/assert"
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
