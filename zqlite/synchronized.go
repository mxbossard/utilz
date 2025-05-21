package zqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mxbossard/utilz/zlog"
	"github.com/mxbossard/utilz/zql"
)

var (
	logger = zlog.New()
)

type SynchronizedDB struct {
	*zql.SynchronizedDB
	busyTimeout time.Duration
}

func sqLiteDbConfigurer(busyTimeout time.Duration) func(db *sql.DB) error {
	return func(db *sql.DB) error {
		// Add SQLite config
		db.SetMaxOpenConns(1)
		db.SetConnMaxIdleTime(busyTimeout * 3)

		// Config to increase DB speed : temp objets and transaction journal stored in memory.
		_, err := db.Exec(fmt.Sprintf(`
			PRAGMA busy_timeout = %d;
			PRAGMA TEMP_STORE = MEMORY;
			PRAGMA JOURNAL_MODE = MEMORY;
			PRAGMA SYNCHRONOUS = OFF;
			PRAGMA LOCKING_MODE = NORMAL;
		`, int64(busyTimeout/time.Millisecond))) // PRAGMA read_uncommitted = true;

		logger.Trace("SynchronizedDB opened")
		return err
	}
}

func OpenSynchronizedDB(backingFile, opts string, busyTimeout time.Duration) (*SynchronizedDB, error) {
	db := zql.NewSynchronizedDB2("sqlite", backingFile, opts, busyTimeout*3, false)
	db.DbConfigurer = sqLiteDbConfigurer(busyTimeout)
	db.IsRetryableError = func(err error) bool {
		return err != nil && strings.Contains(err.Error(), "SQLITE_BUSY")
	}
	db.PanicOnError = true
	return &SynchronizedDB{db, busyTimeout}, nil
}

func OpenSynchronizedClosingDB(backingFile, opts string, busyTimeout time.Duration) (*SynchronizedDB, error) {
	db := zql.NewSynchronizedDB2("sqlite", backingFile, opts, busyTimeout*3, true)
	db.DbConfigurer = sqLiteDbConfigurer(busyTimeout)
	db.IsRetryableError = func(err error) bool {
		return err != nil && strings.Contains(err.Error(), "SQLITE_BUSY")
	}
	db.PanicOnError = true
	return &SynchronizedDB{db, busyTimeout}, nil
}
