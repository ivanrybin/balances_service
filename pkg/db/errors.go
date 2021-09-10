package db

import "fmt"

type InitError struct {
	Info string
	Err  error
}

func (e *InitError) Error() string {
	if e.Info != "" {
		return fmt.Sprintf("db: bad init: %s: %v", e.Info, e.Err)
	}
	return fmt.Sprintf("db: bad init: %v", e.Err)
}

func (e *InitError) Unwrap() error {
	return e.Err
}

type BeginTxError struct {
	Err error
}

func (e *BeginTxError) Error() string {
	return fmt.Sprintf("db: cannot begin transaction: %v", e.Err)
}

func (e *BeginTxError) Unwrap() error {
	return e.Err
}

type CommitTxError struct {
	Err error
}

func (e *CommitTxError) Error() string {
	return fmt.Sprintf("db: cannot commit transaction: %v", e.Err)
}

func (e *CommitTxError) Unwrap() error {
	return e.Err
}

type StmtTxError struct {
	Info string
	Err  error
}

func (e *StmtTxError) Error() string {
	if e.Info != "" {
		return fmt.Sprintf("db: cannot exec statement: %s: %v", e.Info, e.Err)
	}
	return fmt.Sprintf("db: cannot exec statement: %v", e.Err)
}

func (e *StmtTxError) Unwrap() error {
	return e.Err
}

type NoRowsError struct {
	Info string
}

func (e *NoRowsError) Error() string {
	if e.Info != "" {
		return "db: no such row: " + e.Info
	}
	return "db: no such row"
}

type ScanError struct {
	Info string
	Err  error
}

func (e *ScanError) Error() string {
	if e.Info != "" {
		return fmt.Sprintf("db: cannot scan: %s: %v", e.Info, e.Err)
	}
	return fmt.Sprintf("db: cannot scan: %v", e.Err)
}

func (e *ScanError) Unwrap() error {
	return e.Err
}
