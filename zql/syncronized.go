package zql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/flock"
)

type SqlQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type SynchronizedDB struct {
	*sql.DB
	fileLock    *flock.Flock
	busyTimeout time.Duration
}

func (d SynchronizedDB) lock() (err error) {
	lockCtx, cancel := context.WithTimeout(context.Background(), d.busyTimeout)
	defer cancel()
	locked, err := d.fileLock.TryLockContext(lockCtx, time.Millisecond)
	if err != nil {
		return
	}
	if !locked {
		err = errors.New("unable to acquire DB lock")
	}
	return
}

func (d SynchronizedDB) unlock() (err error) {
	if d.fileLock != nil {
		err = d.fileLock.Unlock()
	}
	return
}

func (d SynchronizedDB) Exec(query string, args ...any) (sql.Result, error) {
	d.lock()
	defer d.unlock()
	return d.DB.Exec(query, args...)
}

func (d SynchronizedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.lock()
	defer d.unlock()
	return d.DB.ExecContext(ctx, query, args...)
}

func (d SynchronizedDB) Query(query string, args ...any) (*sql.Rows, error) {
	d.lock()
	defer d.unlock()
	return d.DB.Query(query, args...)
}

func (d SynchronizedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.lock()
	defer d.unlock()
	return d.DB.QueryContext(ctx, query, args...)
}

func (d SynchronizedDB) QueryRow(query string, args ...any) *sql.Row {
	d.lock()
	defer d.unlock()
	return d.DB.QueryRow(query, args...)
}

func (d SynchronizedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	d.lock()
	defer d.unlock()
	return d.DB.QueryRowContext(ctx, query, args...)
}

func (d SynchronizedDB) Begin() (*SynchronizedTx, error) {
	d.lock()
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &SynchronizedTx{Tx: tx, db: &d}, nil
}

func (d SynchronizedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SynchronizedTx, error) {
	d.lock()
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SynchronizedTx{Tx: tx, db: &d}, nil
}

type SynchronizedTx struct {
	*sql.Tx
	db *SynchronizedDB
}

func (t SynchronizedTx) Commit() error {
	defer t.db.unlock()
	return t.Tx.Commit()
}

func (t SynchronizedTx) Rollback() error {
	defer t.db.unlock()
	return t.Tx.Rollback()
}

func NewSynchronizedDB(db *sql.DB, lockingFile, opts string, busyTimeout time.Duration) *SynchronizedDB {
	fileLock := flock.New(lockingFile)
	wrapper := SynchronizedDB{
		DB:          db,
		fileLock:    fileLock,
		busyTimeout: busyTimeout,
	}
	return &wrapper
}
