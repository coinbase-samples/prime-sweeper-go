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

	"github.com/stretchr/testify/assert"
)

func TestFilterWalletsByAssets(t *testing.T) {
	tests := []struct {
		name             string
		assets           []string
		tradingWallets   map[string]core.WalletResponse
		expectedFiltered map[string]core.WalletResponse
	}{
		{
			name:   "single asset match",
			assets: []string{"ETH"},
			tradingWallets: map[string]core.WalletResponse{
				"ETH": {Id: "wallet1", Symbol: "ETH"},
				"BTC": {Id: "wallet2", Symbol: "BTC"},
			},
			expectedFiltered: map[string]core.WalletResponse{
				"wallet1": {Id: "wallet1", Symbol: "ETH"},
			},
		},
		{
			name:   "no asset match",
			assets: []string{"XRP"},
			tradingWallets: map[string]core.WalletResponse{
				"ETH": {Id: "wallet1", Symbol: "ETH"},
				"BTC": {Id: "wallet2", Symbol: "BTC"},
			},
			expectedFiltered: map[string]core.WalletResponse{},
		},
		{
			name:   "multiple asset matches",
			assets: []string{"ETH", "BTC"},
			tradingWallets: map[string]core.WalletResponse{
				"ETH": {Id: "wallet1", Symbol: "ETH"},
				"BTC": {Id: "wallet2", Symbol: "BTC"},
				"LTC": {Id: "wallet3", Symbol: "LTC"},
			},
			expectedFiltered: map[string]core.WalletResponse{
				"wallet1": {Id: "wallet1", Symbol: "ETH"},
				"wallet2": {Id: "wallet2", Symbol: "BTC"},
			},
		},
		{
			name:   "asset not in wallets",
			assets: []string{"ADA"},
			tradingWallets: map[string]core.WalletResponse{
				"ETH": {Id: "wallet1", Symbol: "ETH"},
				"BTC": {Id: "wallet2", Symbol: "BTC"},
				"LTC": {Id: "wallet3", Symbol: "LTC"},
			},
			expectedFiltered: map[string]core.WalletResponse{},
		},
		{
			name:   "empty assets list",
			assets: []string{},
			tradingWallets: map[string]core.WalletResponse{
				"ETH": {Id: "wallet1", Symbol: "ETH"},
				"BTC": {Id: "wallet2", Symbol: "BTC"},
			},
			expectedFiltered: map[string]core.WalletResponse{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := core.FilterWalletsByAssets(tc.assets, tc.tradingWallets)
			assert.Equal(t, tc.expectedFiltered, result)
		})
	}
}
