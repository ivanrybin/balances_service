package bank

import "fmt"

type Config struct {
	AddTxTriesCount      int `yaml:"add_tries_count"`
	WithdrawTxTriesCount int `yaml:"withdraw_tries_count"`
	TransferTxTriesCount int `yaml:"transfer_tries_count"`
}

func (c Config) CheckRestrictions() error {
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
