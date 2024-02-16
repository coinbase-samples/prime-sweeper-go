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
	Name         string `yaml:"name" json:"name"`
	Asset        string `yaml:"asset" json:"asset"`
	Description  string `yaml:"description" json:"description"` // Optional
	Type         string `yaml:"type" json:"type"`
	ColdWalletId string `yaml:"cold_wallet_id" json:"cold_wallet_id"`
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
