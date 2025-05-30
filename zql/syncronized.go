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

const tryLockPeriodMin = 10 * time.Microsecond
const tryLockPeriodMax = 210 * time.Microsecond

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
	lockTimeout        time.Duration
	tryLockPeriod      time.Duration
	maxDurationRetries time.Duration
	driverName         string
	datasourceName     string
	openCloseSync      bool
	dbOpen             bool
	DbConfigurer       func(*sql.DB) error

	// TODO implem vvv
	PanicOnError     bool
	RetryOnError     bool
	IsRetryableError func(error) bool
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

	err = d.open()
	if err != nil {
		return err
	}

	return nil
}

func (d *SynchronizedDB) open() error {
	db, err := sql.Open(d.driverName, d.datasourceName)
	if err != nil {
		return err
	}
	d.DB = db

	if d.DbConfigurer != nil {
		d.DbConfigurer(d.DB)
	}

	return nil
}

func (d *SynchronizedDB) Config() error {
	return nil
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

	lockingTimeout := d.lockTimeout * 3
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

func (d *SynchronizedDB) isRetryableError(err error) bool {
	//return err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sql.ErrConnDone) && !errors.Is(err, sql.ErrTxDone)
	return err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sql.ErrConnDone) && !errors.Is(err, sql.ErrTxDone) && (d.IsRetryableError != nil && d.IsRetryableError(err) || d.IsRetryableError == nil)
}

func (d *SynchronizedDB) processError(f func() error) error {
	err := f()
	retries := 0
	firstErr := err
	if d.RetryOnError && d.isRetryableError(err) {
		// No retries for logic errors
		start := time.Now()
		for err != nil && time.Since(start) < d.maxDurationRetries {
			// if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone) {
			// 	// No retries for logic errors
			// 	break
			// }
			sleepDuration := time.Duration(2000+rand.Intn(10000)) * time.Microsecond
			time.Sleep(sleepDuration)
			err = f()
			retries++
		}

		if d.isRetryableError(err) {
			err = fmt.Errorf("SQL operation errored after %s (%d retries): %w \nfirst error: %w", time.Since(start), retries, err, firstErr)
			panic(err)
		} else {
			fmt.Printf("\n<<>> SQL operation delayed during %s (%d retries). First error: %s ; stack trace:\n%s\n", time.Since(start), retries, firstErr, string(debug.Stack()))
		}
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sql.ErrConnDone) && !errors.Is(err, sql.ErrTxDone) && d.PanicOnError {
		panic(err)
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

	err = d.processError(func() error {
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

	err = d.processError(func() error {
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

	err = d.processError(func() error {
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

	err = d.processError(func() error {
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
	err = d.processError(func() error {
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
	err = d.processError(func() error {
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
	err = r.db.processError(func() error {
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
	err = r.db.processError(func() error {
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

	err = t.db.processError(func() error {
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

	err = t.db.processError(func() error {
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

	err = t.db.processError(func() error {
		r, err = t.Tx.Exec(query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) ExecContext(ctx context.Context, query string, args ...any) (r sql.Result, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.processError(func() error {
		r, err = t.Tx.ExecContext(ctx, query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) Query(query string, args ...any) (r *sql.Rows, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.processError(func() error {
		r, err = t.Tx.Query(query, args...)
		return err
	})

	return
}

func (t SynchronizedTx) QueryContext(ctx context.Context, query string, args ...any) (r *sql.Rows, err error) {
	pt := logger.PerfTimer()
	defer pt.End("query", query)

	err = t.db.processError(func() error {
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

func NewSynchronizedDB(db *sql.DB, lockingFile, opts string, lockTimeout time.Duration) *SynchronizedDB {
	wrapper := SynchronizedDB{
		DB:          db,
		lockFile:    lockingFile,
		lockTimeout: lockTimeout,
		//tryLockPeriod:      time.Duration(rand.Intn(200)+200) * time.Microsecond,
		tryLockPeriod:      time.Duration(rand.Intn(int(tryLockPeriodMax)-int(tryLockPeriodMin)) + int(tryLockPeriodMin)),
		maxDurationRetries: lockTimeout * 2 / 3,
	}
	return &wrapper
}

func NewSynchronizedDB2(driverName, backingFile, opts string, lockTimeout time.Duration, openCloseSync bool) *SynchronizedDB {
	lockFile := backingFile //+ ".lock"
	dataSourceName := backingFile
	if opts != "" {
		dataSourceName += "?" + opts
	}
	wrapper := SynchronizedDB{
		backingFile: backingFile,
		lockFile:    lockFile,
		lockTimeout: lockTimeout,
		// tryLockPeriod:      time.Duration(rand.Intn(200)+200) * time.Microsecond,
		tryLockPeriod:      time.Duration(rand.Intn(int(tryLockPeriodMax)-int(tryLockPeriodMin)) + int(tryLockPeriodMin)),
		maxDurationRetries: lockTimeout * 2 / 3,
		driverName:         driverName,
		datasourceName:     dataSourceName,
		openCloseSync:      openCloseSync,
	}
	return &wrapper
}
