package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ivanrybin/work/avito_bank_service/pkg/config"

	_ "github.com/jackc/pgx/v4/stdlib"

	log "github.com/sirupsen/logrus"
)

type BankDB interface {
	Balance(accountId int) (int64, error)
	Add(accountId int, sum int64) error
	CreateID(accountId int) error
	Withdraw(accountId int, sum int64) error
	Transfer(senderId int, recipientId int, sum int64) error
	IsAccountExist(accountId int) (bool, error)
	Close() error
}

func New(ctx context.Context, cfg config.DBConfig) (BankDB, error) {
	b := &bankDB{}

	if ctx != nil {
		b.ctx, b.cancel = context.WithCancel(ctx)
	} else {
		b.ctx, b.cancel = context.WithCancel(context.Background())
	}

	var err error
	b.db, err = sql.Open("pgx", cfg.ConnectURL())
	if err != nil {
		return nil, &InitError{Info: "cannot open", Err: err}
	}

	err = fmt.Errorf("ping error")
	for i := 1; i <= cfg.ConnTriesCnt; i++ {
		log.Printf("trying to connect to database #%d", i)
		if err = b.db.PingContext(b.ctx); err == nil {
			log.Println("database connection established")
			break
		}
		<-time.After(time.Duration(cfg.ConnTryTime) * time.Second)
	}
	if err != nil {
		return nil, &InitError{
			Info: fmt.Sprintf("cannot connect (tries_cnt=%d, try_time=%d)", cfg.ConnTriesCnt, cfg.ConnTryTime),
			Err:  err,
		}
	}

	b.db.SetMaxOpenConns(cfg.MaxOpenConns)
	b.db.SetMaxIdleConns(cfg.MaxIdleConns)

	return b, nil
}

type bankDB struct {
	ctx    context.Context
	cancel context.CancelFunc

	db *sql.DB
}

func (d *bankDB) Balance(accountId int) (int64, error) {
	var balance int64
	err := d.db.QueryRowContext(d.ctx, "SELECT balance FROM accounts_balances WHERE id = $1", accountId).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, &NoRowsError{Info: fmt.Sprintf("balance: account_id=%d", accountId)}
		}
		return 0, &StmtTxError{Err: err}
	}
	return balance, nil
}

func (d *bankDB) Add(accountId int, sum int64) error {
	_, err := d.db.ExecContext(d.ctx, "UPDATE accounts_balances SET balance = balance + $1 WHERE id = $2", sum, accountId)
	if err != nil {
		return &StmtTxError{Info: fmt.Sprintf("add: account_id=%d", accountId), Err: err}
	}
	return nil
}

func (d *bankDB) Withdraw(accountId int, sum int64) error {
	_, err := d.db.ExecContext(d.ctx, "UPDATE accounts_balances SET balance = balance - $1 WHERE id = $2", sum, accountId)
	if err != nil {
		return &StmtTxError{Info: fmt.Sprintf("withdraw: account_id=%d", accountId), Err: err}
	}
	return nil
}

func (d *bankDB) Transfer(senderId int, recipientId int, sum int64) error {
	tx, err := d.db.BeginTx(d.ctx, nil)
	if err != nil {
		return &BeginTxError{Err: err}
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(d.ctx, "UPDATE accounts_balances SET balance = balance - $1 WHERE id = $2", sum, senderId)
	if err != nil {
		return &StmtTxError{Info: fmt.Sprintf("transfer: sender_id=%d, recipient_id=%d", senderId, recipientId), Err: err}
	}
	_, err = tx.ExecContext(d.ctx, "UPDATE accounts_balances SET balance = balance + $1 WHERE id = $2", sum, recipientId)
	if err != nil {
		return &StmtTxError{Info: fmt.Sprintf("transfer: sender_id=%d, recipient_id=%d", senderId, recipientId), Err: err}
	}

	if err = tx.Commit(); err != nil {
		return &CommitTxError{Err: err}
	}
	return nil
}

func (d *bankDB) CreateID(accountId int) error {
	_, err := d.db.ExecContext(d.ctx, "INSERT INTO accounts_balances VALUES ($1, 0)", accountId)
	if err != nil {
		return &StmtTxError{Info: fmt.Sprintf("create id: account_id=%d", accountId), Err: err}
	}
	return nil
}

func (d *bankDB) IsAccountExist(accountId int) (bool, error) {
	var id int64
	err := d.db.QueryRowContext(d.ctx, "SELECT id FROM accounts_balances WHERE id = $1", accountId).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, &StmtTxError{Info: fmt.Sprintf("is account exist: %d", accountId), Err: err}
	}
	return true, nil
}

func (d *bankDB) Close() error {
	defer d.cancel()
	return d.db.Close()
}
