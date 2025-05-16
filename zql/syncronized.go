package zql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"time"

	"github.com/gofrs/flock"
	"github.com/mxbossard/utilz/zlog"
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
	QueryRow(query string, args ...any) *SynchronizedRow
	QueryRowContext(ctx context.Context, query string, args ...any) *SynchronizedRow
}

type SynchronizedDB struct {
	*sql.DB
	backingFile        string
	lockFile           string
	fileLock           *flock.Flock
	sessionLocked      bool
	busyTimeout        time.Duration
	tryLockPeriod      time.Duration
	maxDurationRetries time.Duration
	driverName         string
	datasourceName     string
	openCloseSync      bool
	dbOpen             bool
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
	db.SetConnMaxIdleTime(time.Second)

	// Config to increase DB speed : temp objets and transaction journal stored in memory.
	_, err = db.Exec(`
		PRAGMA busy_tymeout = 1000;
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
	`) // PRAGMA read_uncommitted = true;
	logger.Trace("SynchronizedDB opened")
	return err
}

func (d *SynchronizedDB) Close() error {
	err := d.DB.Close()
	logger.Trace("SynchronizedDB closed")
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
	perf := logger.TraceTimer()
	defer perf.End()

	// FIXME: do we need to rebuild file lock on every lock ?
	fileLock := flock.New(d.lockFile)

	lockingTimeout := d.busyTimeout * 3
	lockCtx, cancel := context.WithTimeout(context.Background(), lockingTimeout)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, d.tryLockPeriod)
	if err != nil {
		return fmt.Errorf("unable to lock DB (timeout: %s): %w", lockingTimeout, err)
	}
	if !locked {
		err = errors.New("unable to acquire DB lock")
		if err != nil {
			return err
		}
	}
	d.fileLock = fileLock

	if d.openCloseSync || !d.dbOpen {
		// Open DB
		err = d.open()
		if err != nil {
			return err
		}
		d.dbOpen = true
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
		d.dbOpen = false
	}

	if d.fileLock != nil {
		//time.Sleep(100 * time.Millisecond)
		err = d.fileLock.Unlock()
		if err != nil {
			return
		}
		//logger.Debug("SynchronizedDB unlocked ...", "fileLock", d.fileLock)
		d.fileLock = nil
	}

	return
}

func (d *SynchronizedDB) retryOnError(f func() error) error {
	err := f()
	retries := 0
	firstErr := err
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sql.ErrConnDone) && !errors.Is(err, sql.ErrTxDone) {
		// No retries for logic errors
		start := time.Now()
		for err != nil && time.Since(start) < d.maxDurationRetries {
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone) {
				// No retries for logic errors
				break
			}
			sleepDuration := time.Duration(2000+rand.Intn(10000)) * time.Microsecond
			time.Sleep(sleepDuration)
			err = f()
			retries++
		}

		if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sql.ErrConnDone) && !errors.Is(err, sql.ErrTxDone) {
			err = fmt.Errorf("SQL operation errored after %s (%d retries): %w \nfirst error: %w", time.Since(start), retries, err, firstErr)
			panic(err)
		} else {
			fmt.Printf("\n<<>> SQL operation delayed during %s (%d retries). First error: %s ; stack trace:\n%s\n", time.Since(start), retries, firstErr, string(debug.Stack()))
		}
	}

	return err
}

func (d *SynchronizedDB) Exec(query string, args ...any) (r sql.Result, err error) {
	err = d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = d.retryOnError(func() error {
		r, err = d.DB.Exec(query, args...)
		return err
	})

	return
}

func (d *SynchronizedDB) ExecContext(ctx context.Context, query string, args ...any) (r sql.Result, err error) {
	err = d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = d.retryOnError(func() error {
		r, err = d.DB.ExecContext(ctx, query, args...)
		return err
	})

	return
}

func (d *SynchronizedDB) Query(query string, args ...any) (r *sql.Rows, err error) {
	err = d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = d.retryOnError(func() error {
		r, err = d.DB.Query(query, args...)
		return err
	})

	return
}

func (d *SynchronizedDB) QueryContext(ctx context.Context, query string, args ...any) (r *sql.Rows, err error) {
	err = d.lock(false)
	if err != nil {
		return nil, err
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = d.retryOnError(func() error {
		r, err = d.DB.QueryContext(ctx, query, args...)
		return err
	})

	return
}

func (d *SynchronizedDB) QueryRow(query string, args ...any) *SynchronizedRow {
	err := d.lock(false)
	if err != nil {
		panic(err)
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	return &SynchronizedRow{Row: d.DB.QueryRow(query, args...), db: d}
}

func (d *SynchronizedDB) QueryRowContext(ctx context.Context, query string, args ...any) *SynchronizedRow {
	err := d.lock(false)
	if err != nil {
		panic(err)
	}
	defer d.unlock(false)
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	return &SynchronizedRow{Row: d.DB.QueryRowContext(ctx, query, args...), db: d}
}

func (d *SynchronizedDB) Begin() (*SynchronizedTx, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}

	var t *sql.Tx
	err = d.retryOnError(func() error {
		t, err = d.DB.Begin()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &SynchronizedTx{Tx: t, db: d}, nil
}

func (d *SynchronizedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*SynchronizedTx, error) {
	err := d.lock(false)
	if err != nil {
		return nil, err
	}

	var t *sql.Tx
	err = d.retryOnError(func() error {
		t, err = d.DB.BeginTx(ctx, opts)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &SynchronizedTx{Tx: t, db: d}, nil
}

type SynchronizedRow struct {
	*sql.Row
	db *SynchronizedDB
}

func (r *SynchronizedRow) Scan(dest ...any) (err error) {
	err = r.db.retryOnError(func() error {
		err = r.Row.Scan(dest...)
		return err
	})
	return
}

type SynchronizedRows struct {
	*sql.Rows
	db *SynchronizedDB
}

func (r *SynchronizedRows) Scan(dest ...any) (err error) {
	err = r.db.retryOnError(func() error {
		err = r.Rows.Scan(dest...)
		return err
	})
	return
}

type SynchronizedTx struct {
	*sql.Tx
	db *SynchronizedDB
}

func (t SynchronizedTx) Commit() (err error) {
	defer t.db.unlock(false)

	err = t.db.retryOnError(func() error {
		err = t.Tx.Commit()
		return err
	})

	if errors.Is(err, sql.ErrTxDone) {
		// Suppress TX done error on commit
		err = nil
	}

	return
}

func (t SynchronizedTx) Rollback() (err error) {
	defer t.db.unlock(false)

	err = t.db.retryOnError(func() error {
		err = t.Tx.Rollback()
		return err
	})

	if errors.Is(err, sql.ErrTxDone) {
		// Suppress TX done error on rollback
		err = nil
	}

	return
}

func (t SynchronizedTx) Exec(query string, args ...any) (r sql.Result, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.retryOnError(func() error {
		r, err = t.Tx.Exec(query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) ExecContext(ctx context.Context, query string, args ...any) (r sql.Result, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.retryOnError(func() error {
		r, err = t.Tx.ExecContext(ctx, query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) Query(query string, args ...any) (r *sql.Rows, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.retryOnError(func() error {
		r, err = t.Tx.Query(query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) QueryContext(ctx context.Context, query string, args ...any) (r *sql.Rows, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.retryOnError(func() error {
		r, err = t.Tx.QueryContext(ctx, query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) QueryRow(query string, args ...any) *SynchronizedRow {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	return &SynchronizedRow{Row: t.Tx.QueryRow(query, args...), db: t.db}
}

func (t SynchronizedTx) QueryRowContext(ctx context.Context, query string, args ...any) *SynchronizedRow {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	return &SynchronizedRow{Row: t.Tx.QueryRowContext(ctx, query, args...), db: t.db}
}

func NewSynchronizedDB(db *sql.DB, lockingFile, opts string, busyTimeout time.Duration) *SynchronizedDB {
	wrapper := SynchronizedDB{
		DB:                 db,
		lockFile:           lockingFile,
		busyTimeout:        busyTimeout,
		tryLockPeriod:      time.Duration(rand.Intn(200) * int(time.Microsecond)),
		maxDurationRetries: busyTimeout * 2,
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
		backingFile:        backingFile,
		lockFile:           lockFile,
		busyTimeout:        busyTimeout,
		tryLockPeriod:      time.Duration(100+rand.Intn(200)) * time.Microsecond,
		maxDurationRetries: busyTimeout * 2,
		driverName:         driverName,
		datasourceName:     dataSourceName,
		openCloseSync:      openCloseSync,
	}
	return &wrapper
}
