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

package model

import (
	"time"
)

type Config struct {
	Daemon  DaemonConfig `yaml:"daemon"`
	Rules   []Rule       `yaml:"rules"`
	Wallets []Wallet     `yaml:"wallets"`
}

type DaemonConfig struct {
	ContextTimeoutDuration         int           `yaml:"context_timeout_duration"`
	TransferMonitorFrequency       time.Duration `yaml:"transfer_monitor_frequency"`
	TransferMonitorTimeoutDuration time.Duration `yaml:"transfer_monitor_timeout_duration"`
}

type Rule struct {
	Direction   string   `yaml:"direction" json:"direction"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"` // Optional
	Schedule    string   `yaml:"schedule" json:"schedule"`
	Wallets     []string `yaml:"wallets" json:"wallets"`
}

type Wallet struct {
	Name        string `yaml:"name" json:"name"`
	Asset       string `yaml:"asset" json:"asset"`
	Description string `yaml:"description" json:"description"` // Optional
	Type        string `yaml:"type" json:"type"`
	WalletId    string `yaml:"wallet_id" json:"wallet_id"`
}

const (
	HotToCold TransferDirection = "trading_to_cold_custody"
	ColdToHot TransferDirection = "cold_custody_to_trading"
)

type TransferDirection string

type TransferDetails struct {
	Direction   TransferDirection
	WalletNames []string
	OperationId string
	RuleName    string
}
