package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_ConfigDecoder(t *testing.T) {
	dbCfg := DBConfig{
		Name:     "db",
		Host:     "host",
		Port:     5432,
		User:     "user",
		Password: "password",
	}
	serverCfg := ServerConfig{
		Port: 12345,
		Host: "host",
	}
	bankCfg := BankConfig{
		AddTxTriesCount:      1,
		TransferTxTriesCount: 1,
		WithdrawTxTriesCount: 1,
	}
	cfg := Config{
		DB:     dbCfg,
		Server: serverCfg,
		Bank:   bankCfg,
	}

	buf := &bytes.Buffer{}
	e := yaml.NewEncoder(buf)
	assert.Nil(t, e.Encode(cfg))

	d := yaml.NewDecoder(buf)
	cfgDecoded := &Config{}
	assert.Nil(t, d.Decode(cfgDecoded))

	assert.Equal(t, cfg, *cfgDecoded, "configs not equal")
}
