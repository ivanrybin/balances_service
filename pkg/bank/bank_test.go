package bank

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"ivanrybin/work/avito_bank_service/pkg/config"
	"ivanrybin/work/avito_bank_service/pkg/db"
)

type dbMock struct {
	m       sync.Mutex
	balance map[int]int64
}

func (d *dbMock) CreateID(accountId int) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.balance[accountId] = 0
	return nil
}

func (d *dbMock) Balance(accountId int) (int64, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if ok, _ := d.IsAccountExist(accountId); !ok {
		return 0, &db.NoRowsError{}
	}
	return d.balance[accountId], nil
}

func (d *dbMock) Add(accountId int, sum int64) error {
	d.m.Lock()
	defer d.m.Unlock()
	if ok, _ := d.IsAccountExist(accountId); !ok {
		return &db.NoRowsError{}
	}
	d.balance[accountId] += sum
	return nil
}

func (d *dbMock) Withdraw(accountId int, sum int64) error {
	d.m.Lock()
	defer d.m.Unlock()
	if ok, _ := d.IsAccountExist(accountId); !ok {
		return &db.NoRowsError{}
	}
	d.balance[accountId] -= sum
	return nil
}

func (d *dbMock) Transfer(senderId int, recipientId int, sum int64) error {
	d.m.Lock()
	defer d.m.Unlock()
	if ok, _ := d.IsAccountExist(senderId); !ok {
		return &db.NoRowsError{}
	}
	if ok, _ := d.IsAccountExist(recipientId); !ok {
		return &db.NoRowsError{}
	}
	d.balance[senderId] -= sum
	d.balance[recipientId] += sum
	return nil
}

func (d *dbMock) IsAccountExist(accountId int) (bool, error) {
	_, ok := d.balance[accountId]
	return ok, nil
}

func (d *dbMock) Close() error {
	return nil
}

func NewDB() *dbMock {
	return &dbMock{balance: map[int]int64{}}
}

func TestBank_Basics(t *testing.T) {
	_db := NewDB()
	cfg := config.BankConfig{AddTxTriesCount: 1, WithdrawTxTriesCount: 1, TransferTxTriesCount: 1, RatesAPIToken: "Token"}

	b, err := New(context.Background(), _db, cfg)
	assert.Nil(t, err)

	centSum := int64(10)

	err = b.Add(1, centSum)
	assert.Nil(t, err)

	balance, err := b.Balance(1, "RUB")
	assert.Nil(t, err)
	assert.Equal(t, centSum, balance)

	_, err = b.Balance(42, "RUB")
	assert.Nil(t, err)
	assert.NotNil(t, err)
}

func TestBank_WithDraw(t *testing.T) {
	_db := NewDB()
	cfg := config.BankConfig{AddTxTriesCount: 1, WithdrawTxTriesCount: 1, TransferTxTriesCount: 1}

	b, err := New(context.Background(), _db, cfg)
	assert.Nil(t, err)

	centSum := int64(10)

	err = b.Add(1, centSum)
	assert.Nil(t, err)

	err = b.Withdraw(1, centSum*10)
	assert.NotNil(t, err)

	balance, err := b.Balance(1, "RUB")
	assert.Nil(t, err)
	assert.Equal(t, centSum, balance)

	err = b.Withdraw(1, centSum/2)
	assert.Nil(t, err)

	balance, err = b.Balance(1, "RUB")
	assert.Nil(t, err)
	assert.Equal(t, centSum/2, balance)
}

func TestBank_Transfer(t *testing.T) {
	_db := NewDB()
	cfg := config.BankConfig{AddTxTriesCount: 1, WithdrawTxTriesCount: 1, TransferTxTriesCount: 1}

	b, err := New(context.Background(), _db, cfg)
	assert.Nil(t, err)

	b1 := int64(200)
	b2 := int64(50)

	sender, recipient := 1, 2

	_db.balance[1] = b1
	_db.balance[2] = b2

	err = b.Transfer(sender, recipient, 1000000)
	assert.NotNil(t, err)

	for i := 0; i < 10; i++ {
		b1 -= 100
		b2 += 100

		err = b.Transfer(sender, recipient, 100)
		assert.Nil(t, err)

		balance, err := b.Balance(sender, "RUB")
		assert.Nil(t, err)
		assert.Equal(t, b1, balance)

		balance, err = b.Balance(recipient, "RUB")
		assert.Nil(t, err)
		assert.Equal(t, b2, balance)

		sender, recipient = recipient, sender
		b1, b2 = b2, b1
	}
}
