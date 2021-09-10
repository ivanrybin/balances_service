package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Bank   BankConfig   `yaml:"bank"`
	DB     DBConfig     `yaml:"database"`
}

func Load(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer func() { _ = f.Close() }()

	c := Config{}
	d := yaml.NewDecoder(f)
	if err = d.Decode(&c); err != nil {
		return Config{}, err
	}
	return c, nil
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type DBConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	ConnTryTime  int `yaml:"conn_try_time"`
	ConnTriesCnt int `yaml:"conn_tries_cnt"`

	MaxIdleConns int `yaml:"max_idle_conns"`
	MaxOpenConns int `yaml:"max_open_conns"`
}

func (c *DBConfig) ConnectURL() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + strconv.Itoa(c.Port) + "/" + c.Name
}

type BankConfig struct {
	AddTxTriesCount      int `yaml:"add_tries_count"`
	WithdrawTxTriesCount int `yaml:"withdraw_tries_count"`
	TransferTxTriesCount int `yaml:"transfer_tries_count"`

	RatesAPIToken string `yaml:"rates_api_token"`
}

func (c BankConfig) CheckRestrictions() error {
	switch {
	case c.AddTxTriesCount <= 0:
		return fmt.Errorf("add tx tries count <= 0: %d", c.AddTxTriesCount)
	case c.WithdrawTxTriesCount <= 0:
		return fmt.Errorf("withdraw tx tries count <= 0: %d", c.WithdrawTxTriesCount)
	case c.TransferTxTriesCount <= 0:
		return fmt.Errorf("transfer tx tries count <= 0: %d", c.TransferTxTriesCount)
	}
	return nil
}
