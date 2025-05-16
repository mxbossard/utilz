package zqlite

import (
	"time"

	"github.com/mxbossard/utilz/zql"
)

func OpenSynchronizedDB(backingFile, opts string, busyTimeout time.Duration) (*zql.SynchronizedDB, error) {
	return zql.NewSynchronizedDB2("sqlite", backingFile, opts, busyTimeout, false), nil
}

func OpenSynchronizedClosingDB(backingFile, opts string, busyTimeout time.Duration) (*zql.SynchronizedDB, error) {
	return zql.NewSynchronizedDB2("sqlite", backingFile, opts, busyTimeout, true), nil
}
