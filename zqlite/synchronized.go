package zqlite

import (
	"database/sql"
	"time"

	"mby.fr/utils/zql"
)

func OpenSynchronizedDB(backingFile, opts string, busyTimeout time.Duration) (*zql.SynchronizedDB, error) {
	dataSourceName := backingFile
	if opts != "" {
		dataSourceName += "?" + opts
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	return zql.NewSynchronizedDB(db, backingFile, opts, busyTimeout), nil
}
