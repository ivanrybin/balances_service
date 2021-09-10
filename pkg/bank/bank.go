package bank

import (
	"context"
	"fmt"

	"ivanrybin/work/avito_bank_service/pkg/config"
	"ivanrybin/work/avito_bank_service/pkg/currconv"
	"ivanrybin/work/avito_bank_service/pkg/db"

	log "github.com/sirupsen/logrus"
)

type Bank interface {
	Balance(accountId int, currency string) (int64, error)
	Add(accountId int, centsSum int64) error
	Withdraw(accountId int, centsSum int64) error
	Transfer(senderId int, recipientId int, centsSum int64) error
}

func New(ctx context.Context, db db.BankDB, cfg config.BankConfig) (Bank, error) {
	var err error
	if err = cfg.CheckRestrictions(); err != nil {
		return nil, fmt.Errorf("invalid config")
	}

	b := &bank{
		cfg: cfg,
		db:  db,
	}

	if ctx != nil {
		b.ctx, b.cancel = context.WithCancel(ctx)
	} else {
		b.ctx, b.cancel = context.WithCancel(context.Background())
	}

	b.converter, err = currconv.New(b.ctx, cfg.RatesAPIToken)
	if err != nil {
		return nil, fmt.Errorf("cannot init currency converter")
	}

	return b, nil
}

type bank struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg config.BankConfig

	db        db.BankDB
	converter *currconv.Converter
}

func (b *bank) Balance(accountId int, currency string) (int64, error) {
	if ok, err := b.checkAccountExist(accountId); err != nil {
		return 0, operationErr("balance", err)
	} else if !ok {
		return 0, &NoAccountError{}
	}

	balance, err := b.db.Balance(accountId)
	if err != nil {
		return 0, operationErr("balance", err)
	}

	convertedBalance, err := b.converter.FromRUB(balance, currency)
	if err != nil {
		return 0, operationErr("balance", err)
	}
	return convertedBalance, nil
}

func (b *bank) Add(accountId int, centsSum int64) error {
	switch {
	case centsSum < 0:
		return &NegativeSumError{Method: "add"}
	case centsSum == 0:
		return &ZeroSumError{Method: "add"}
	}
	return b.add(accountId, centsSum)
}

func (b *bank) add(accountId int, sum int64) error {
	if ok, err := b.checkAccountExist(accountId); err != nil {
		return operationErr("add", err)
	} else if !ok {
		if err = b.db.CreateID(accountId); err != nil {
			return operationErr("add", err)
		}
	}

	// try to add N times (transaction may fail)
	var transErr error
	for i := 0; i < b.cfg.AddTxTriesCount; i++ {
		if transErr = b.db.Add(accountId, sum); transErr == nil {
			return nil
		} else {
			log.Warnln(transErr)
		}
	}
	return operationErr("add", transErr)
}

func (b *bank) Withdraw(accountId int, centsSum int64) error {
	switch {
	case centsSum < 0:
		return &NegativeSumError{Method: "withdraw"}
	case centsSum == 0:
		return &ZeroSumError{Method: "withdraw"}
	}
	return b.withdraw(accountId, centsSum)
}

func (b *bank) withdraw(accountId int, sum int64) error {
	// check account exist
	if ok, err := b.checkAccountExist(accountId); err != nil {
		return operationErr("transfer", err)
	} else if !ok {
		return &NoAccountError{}
	}

	// check balance limit
	if ok, err := b.checkEnoughBalanceToWithdraw(accountId, sum); err != nil {
		return operationErr("transfer", err)
	} else if !ok {
		return &NotEnoughMoneyError{}
	}

	// try to withdraw N times (transaction may fail)
	var transErr error
	for i := 0; i < b.cfg.WithdrawTxTriesCount; i++ {

		// withdraw transaction
		if transErr = b.db.Withdraw(accountId, sum); transErr == nil {
			return nil
		} else {
			log.Warnln(transErr)
		}

		// balance may have changed during transaction: check balance again
		if ok, err := b.checkEnoughBalanceToWithdraw(accountId, sum); err != nil {
			return operationErr("withdraw", err)
		} else if !ok {
			return &NotEnoughMoneyError{}
		}

	}
	return operationErr("withdraw", transErr)
}

func (b *bank) Transfer(senderId int, recipientId int, centsSum int64) error {
	switch {
	case centsSum < 0:
		return &NegativeSumError{Method: "transfer"}
	case centsSum == 0:
		return &ZeroSumError{Method: "transfer"}
	}
	return b.transfer(senderId, recipientId, centsSum)
}

func (b *bank) transfer(senderId int, recipientId int, sum int64) error {
	// check sender account exist
	if ok, err := b.checkAccountExist(senderId); err != nil {
		return operationErr("transfer", err)
	} else if !ok {
		return &NoAccountError{"sender"}
	}

	// check recipient account exist
	if ok, err := b.checkAccountExist(recipientId); err != nil {
		return operationErr("transfer", err)
	} else if !ok {
		return &NoAccountError{"recipient"}
	}

	// check sender balance limit
	if ok, err := b.checkEnoughBalanceToWithdraw(senderId, sum); err != nil {
		return operationErr("transfer", err)
	} else if !ok {
		return &NotEnoughMoneyError{}
	}

	// try to transfer money N times (transaction may fail: deadlock / something else)
	var transErr error
	for i := 0; i < b.cfg.TransferTxTriesCount; i++ {

		// transfer transaction
		if transErr = b.db.Transfer(senderId, recipientId, sum); transErr == nil {
			return nil
		} else {
			log.Warnln(transErr)
		}

		// senders balance may have changed during transaction: check balance again
		if ok, err := b.checkEnoughBalanceToWithdraw(senderId, sum); err != nil {
			return operationErr("transfer", err)
		} else if !ok {
			return &NotEnoughMoneyError{}
		}

	}
	return operationErr("transfer", transErr)
}

func (b *bank) checkAccountExist(accountId int) (bool, error) {
	isExist, err := b.db.IsAccountExist(accountId)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (b *bank) checkEnoughBalanceToWithdraw(accountId int, sum int64) (bool, error) {
	balance, err := b.db.Balance(accountId)
	if err != nil {
		return false, err
	}
	return sum <= balance, nil
}

func operationErr(op string, err error) error {
	log.Warnln(err)
	return &OperationError{Op: op}
}
