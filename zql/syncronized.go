package zql

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/gofrs/flock"
	"mby.fr/utils/zlog"
)

var (
	// logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	// 	Level: slog.LevelWarn,
	// }))
	logger = zlog.New()
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
	backingFile    string
	lockFile       string
	fileLock       *flock.Flock
	sessionLocked  bool
	busyTimeout    time.Duration
	driverName     string
	datasourceName string
	openCloseSync  bool
}

func (d *SynchronizedDB) FileLockPath() string {
	return d.lockFile
}

func (d *SynchronizedDB) Open() error {
	err := d.lock(false)
	if err != nil {
		return err
	}
	defer d.unlock(false)

	return d.open()
}

func (d *SynchronizedDB) open() error {
	db, err := sql.Open(d.driverName, d.datasourceName)
	if err != nil {
		return err
	}
	d.DB = db

	// FIXME: Add Configrable pragma
	db.SetMaxOpenConns(5)

	// Config to increase DB speed : temp objets and transaction journal stored in memory.
	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
	`)
	logger.Debug("SynchronizedDB opened")
	return err
}

func (d *SynchronizedDB) Close() error {
	err := d.DB.Close()
	logger.Debug("SynchronizedDB closed")
	return err
}

func (d *SynchronizedDB) Remove() error {
	err := os.Remove(d.lockFile)
	if err != nil {
		return err
	}
	err = os.Remove(d.backingFile)
	return err
}

func (d *SynchronizedDB) Lock() error {
	err := d.lock(true)
	if err != nil {
		return err
	}
	d.sessionLocked = true
	logger.Trace("SynchronizedDB session locked")
	return err
}

func (d *SynchronizedDB) Unlock() error {
	d.sessionLocked = false
	err := d.unlock(true)
	logger.Trace("SynchronizedDB session unlocked")

	return err
}

func (d *SynchronizedDB) lock(session bool) (err error) {
	if d.sessionLocked && !session {
		// Locking is already done at session level
		logger.Trace("SynchronizedDB session already locked")
		return
	}
	//logger.Debug("SynchronizedDB locking ...", "fileLock", d.fileLock)
	perf := logger.PerfTimer()
	defer perf.End()

	fileLock := flock.New(d.lockFile)

	lockCtx, cancel := context.WithTimeout(context.Background(), d.busyTimeout)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Millisecond)
	if err != nil {
		return
	}
	if !locked {
		err = errors.New("unable to acquire DB lock")
		if err != nil {
			return err
		}
	}
	logger.Perf("just acquired the lock", "lock duration", perf.SinceStart())
	d.fileLock = fileLock

	if d.openCloseSync {
		// Open DB
		err = d.open()
		if err != nil {
			return err
		}
	}

	//logger.Debug("SynchronizedDB locked ...", "fileLock", d.fileLock)
	return
}

func (d *SynchronizedDB) unlock(session bool) (err error) {
	if d.sessionLocked && !session {
		// Locking is already done at session level
		logger.Trace("SynchronizedDB session still locked")
		return
	}
	//logger.Debug("SynchronizedDB unlocking ...", "fileLock", d.fileLock)
	perf := logger.TraceTimer()
	defer perf.End()

	if d.openCloseSync {
		// CLose DB
		err = d.Close()
		if err != nil {
			return err
		}
	}

	if d.fileLock != nil {
		err = d.fileLock.Unlock()
		if err != nil {
			return
		}
		//logger.Debug("SynchronizedDB unlocked ...", "fileLock", d.fileLock)
		d.fileLock = nil
	}

	return
}

func (d *SynchronizedDB) Exec(query string, args ...any) (sql.Result, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)
	return d.DB.Exec(query, args...)
}

func (d *SynchronizedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	return d.DB.ExecContext(ctx, query, args...)
}

func (d *SynchronizedDB) Query(query string, args ...any) (*sql.Rows, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	return d.DB.Query(query, args...)
}

func (d *SynchronizedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	return d.DB.QueryContext(ctx, query, args...)
}

func (d *SynchronizedDB) QueryRow(query string, args ...any) *sql.Row {
	err := d.lock(false)
	if err != nil {
		panic(err)
	}
	defer d.unlock(false)
	return d.DB.QueryRow(query, args...)
}

func (d *SynchronizedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	err := d.lock(false)
	if err != nil {
		panic(err)
	}
	defer d.unlock(false)
	return d.DB.QueryRowContext(ctx, query, args...)
}

func (d *SynchronizedDB) Begin() (*SynchronizedTx, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	tx, err := d.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &SynchronizedTx{Tx: tx, db: d}, nil
}

func (d *SynchronizedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SynchronizedTx, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SynchronizedTx{Tx: tx, db: d}, nil
}

type SynchronizedTx struct {
	*sql.Tx
	db *SynchronizedDB
}

func (t SynchronizedTx) Commit() error {
	defer t.db.unlock(false)
	return t.Tx.Commit()
}

func (t SynchronizedTx) Rollback() error {
	defer t.db.unlock(false)
	return t.Tx.Rollback()
}

func NewSynchronizedDB(db *sql.DB, lockingFile, opts string, busyTimeout time.Duration) *SynchronizedDB {
	wrapper := SynchronizedDB{
		DB:          db,
		lockFile:    lockingFile,
		busyTimeout: busyTimeout,
	}
	return &wrapper
}

func NewSynchronizedDB2(driverName, backingFile, opts string, busyTimeout time.Duration, openCloseSync bool) *SynchronizedDB {
	lockFile := backingFile //+ ".lock"
	dataSourceName := backingFile
	if opts != "" {
		dataSourceName += "?" + opts
	}
	wrapper := SynchronizedDB{
		backingFile:    backingFile,
		lockFile:       lockFile,
		busyTimeout:    busyTimeout,
		driverName:     driverName,
		datasourceName: dataSourceName,
		openCloseSync:  openCloseSync,
	}
	return &wrapper
}
