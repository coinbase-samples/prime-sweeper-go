package model

type Config struct {
	Daemon  DaemonConfig `yaml:"daemon"`
	Rules   []Rule       `yaml:"rules"`
	Wallets []Wallet     `yaml:"wallets"`
}

type DaemonConfig struct {
	TimeoutDuration int `yaml:"timeoutDuration"`
}

type Rule struct {
	Direction   string   `yaml:"direction"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"` // Optional
	Schedule    string   `yaml:"schedule"`
	Wallets     []string `yaml:"wallets"`
}

type Wallet struct {
	Name         string `yaml:"name"`
	Asset        string `yaml:"asset"`
	Description  string `yaml:"description"` // Optional
	Type         string `yaml:"type"`
	ColdWalletId string `yaml:"cold-wallet-id"`
}
