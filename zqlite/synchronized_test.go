package zqlite

import (
	"io"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	_ "modernc.org/sqlite"

	"github.com/gofrs/flock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/zlog"
)

func TestMain(m *testing.M) {
	// test context initialization here
	zlog.ColoredConfig()
	zlog.SetLogLevelThreshold(zlog.LevelTrace)
	zlog.PerfTimerStartAsTrace(false)
	os.Exit(m.Run())
}

func TestMutex(t *testing.T) {
	fl := sync.Mutex{}

	count := 10
	epsilon := 0
	var wg sync.WaitGroup
	for k := 0; k < count; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fl.Lock()
			defer fl.Unlock()
			epsilon++
			// Epsilon should always be 1
			require.Equal(t, 1, epsilon)
			time.Sleep(2 * time.Millisecond)
			epsilon--

		}()
	}
	wg.Wait()

	require.Equal(t, 0, epsilon)
}

func TestUnixFlock(t *testing.T) {
	// t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestUnixFlock")
	//m := &sync.Mutex{}

	count := 10
	epsilon := 0
	var wg sync.WaitGroup
	for k := 0; k < count; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file := filez.Open3OrPanic(filepath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
			//m.Lock()
			err := unix.Flock(int(file.Fd()), unix.LOCK_EX)
			assert.NoError(t, err)
			//m.Unlock()
			defer func() {
				err := unix.Flock(int(file.Fd()), unix.LOCK_UN)
				require.NoError(t, err)
			}()
			epsilon++
			// Epsilon should always be 1
			require.Equal(t, 1, epsilon)
			time.Sleep(2 * time.Millisecond)
			_, err = file.WriteString("foo")
			require.NoError(t, err)
			epsilon--

		}()
	}
	wg.Wait()

	require.Equal(t, 0, epsilon)
}

func TestFcntlFlock(t *testing.T) {
	t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestFcntlFlock")
	file := filez.Open3OrPanic(filepath, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0666)
	//m := &sync.Mutex{}

	count := 10
	epsilon := 0
	var wg sync.WaitGroup
	for k := 0; k < count; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			flockT := syscall.Flock_t{
				Type:   syscall.F_WRLCK,
				Whence: io.SeekStart,
				Start:  0,
				Len:    0,
			}
			//m.Lock()
			err := syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, &flockT)
			assert.NoError(t, err)
			//m.Unlock()
			defer func() {
				err := syscall.FcntlFlock(file.Fd(), syscall.LOCK_UN, &flockT)
				require.NoError(t, err)
			}()
			epsilon++
			// Epsilon should always be 1
			require.Equal(t, 1, epsilon)
			time.Sleep(2 * time.Millisecond)
			_, err = file.WriteString("foo")
			require.NoError(t, err)
			epsilon--

		}()
	}
	wg.Wait()

	require.Equal(t, 0, epsilon)
}

func TestFlock(t *testing.T) {
	filepath := filez.MkTempOrPanic("zqlite-TestFlock")
	//filepath = "/var/lock/flock.lock"
	os.Remove(filepath)
	defer os.Remove(filepath)
	//m := &sync.Mutex{}

	count := 10
	epsilon := 0
	var wg sync.WaitGroup
	for k := 0; k < count; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fl := flock.New(filepath)
			//m.Lock()
			err := fl.Lock()
			require.NoError(t, err)
			//m.Unlock()
			defer fl.Unlock()
			epsilon++
			// Epsilon should always be 1
			require.Equal(t, 1, epsilon)
			time.Sleep(2 * time.Millisecond)
			epsilon--

		}()
	}
	wg.Wait()

	require.Equal(t, 0, epsilon)
}

func TestSimultaneousUsage_OneOpening_OneConn(t *testing.T) {
	t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestSimultaneousUsage")

	db, err := OpenSynchronizedDB(filepath, "", 2*time.Second)
	assert.NoError(t, err)
	require.NotNil(t, db)
	defer db.Remove()

	err = db.Open()
	require.NoError(t, err)

	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
		CREATE TABLE IF NOT EXISTS foo (
			seq INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO foo(seq) VALUES (0);
	`)
	assert.NoError(t, err)

	expectedSeq := 50
	var wg sync.WaitGroup
	for k := 0; k < expectedSeq; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = db.Exec(`UPDATE foo SET seq = (SELECT seq + 1 FROM foo);`)
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	var seq int
	row := db.QueryRow(`SELECT seq FROM foo;`)
	row.Scan(&seq)

	assert.Equal(t, expectedSeq, seq)

	err = db.Close()
	require.NoError(t, err)
}

func TestSimultaneousUsage_OneOpening_TenConns(t *testing.T) {
	//t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestSimultaneousUsage_OneOpening_TenConns")

	db, err := OpenSynchronizedDB(filepath, "", 2*time.Second)
	assert.NoError(t, err)
	require.NotNil(t, db)
	defer db.Remove()

	err = db.Open()
	require.NoError(t, err)

	db.SetMaxOpenConns(10)

	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
		CREATE TABLE IF NOT EXISTS foo (
			seq INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO foo(seq) VALUES (0);
	`)
	assert.NoError(t, err)

	expectedSeq := 50
	var wg sync.WaitGroup
	for k := 0; k < expectedSeq; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = db.Exec(`UPDATE foo SET seq = (SELECT seq + 1 FROM foo);`)
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	var seq int
	row := db.QueryRow(`SELECT seq FROM foo;`)
	row.Scan(&seq)

	assert.Equal(t, expectedSeq, seq)

	err = db.Close()
	require.NoError(t, err)
}

func TestSimultaneousUsage_OneOpening_TenConns_Closing(t *testing.T) {
	//t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestSimultaneousUsage_OneOpening_TenConns")

	db, err := OpenSynchronizedClosingDB(filepath, "", 2*time.Second)
	assert.NoError(t, err)
	require.NotNil(t, db)
	defer db.Remove()

	// err = db.Open()
	// require.NoError(t, err)

	// db.SetMaxOpenConns(10)

	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
		CREATE TABLE IF NOT EXISTS foo (
			seq INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO foo(seq) VALUES (0);
	`)
	assert.NoError(t, err)

	expectedSeq := 12
	var wg sync.WaitGroup
	for k := 0; k < expectedSeq; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err = db.Exec(`UPDATE foo SET seq = (SELECT seq + 1 FROM foo);`)
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	var seq int
	row := db.QueryRow(`SELECT seq FROM foo;`)
	row.Scan(&seq)

	assert.Equal(t, expectedSeq, seq)

	err = db.Close()
	require.NoError(t, err)
}

func TestSimultaneousUsage_OneDbByQuery(t *testing.T) {
	//t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestSimultaneousUsage_OneDbByQuery")

	db, err := OpenSynchronizedDB(filepath, "", 2*time.Second)
	assert.NoError(t, err)
	require.NotNil(t, db)
	defer db.Remove()

	err = db.Open()
	require.NoError(t, err)

	//db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
		CREATE TABLE IF NOT EXISTS foo (
			seq INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO foo(seq) VALUES (0);
	`)
	assert.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	expectedSeq := 50
	var wg sync.WaitGroup
	for k := 0; k < expectedSeq; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			db, err := OpenSynchronizedDB(filepath, "", 2*time.Second)
			assert.NoError(t, err)
			require.NotNil(t, db)

			err = db.Open()
			require.NoError(t, err)

			//db.SetMaxOpenConns(1)

			_, err = db.Exec(`UPDATE foo SET seq = (SELECT seq + 1 FROM foo);`)
			require.NoError(t, err)

			err = db.Close()
			require.NoError(t, err)

		}()
	}
	wg.Wait()

	err = db.Open()
	require.NoError(t, err)

	var seq int
	row := db.QueryRow(`SELECT seq FROM foo;`)
	row.Scan(&seq)

	assert.Equal(t, expectedSeq, seq)

	err = db.Close()
	require.NoError(t, err)
}

func TestSimultaneousUsage_OpenCloseEveryTime(t *testing.T) {
	//t.Skip()
	filepath := filez.MkTempOrPanic("zqlite-TestSimultaneousUsage_OpenCloseEveryTime")

	db, err := OpenSynchronizedDB(filepath, "", 2*time.Second)
	assert.NoError(t, err)
	require.NotNil(t, db)
	defer db.Remove()

	err = db.Open()
	require.NoError(t, err)

	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		PRAGMA TEMP_STORE = MEMORY;
		PRAGMA JOURNAL_MODE = MEMORY;
		PRAGMA SYNCHRONOUS = OFF;
		PRAGMA LOCKING_MODE = NORMAL;
		CREATE TABLE IF NOT EXISTS foo (
			seq INTEGER NOT NULL DEFAULT 0
		);
		INSERT INTO foo(seq) VALUES (0);
	`)
	assert.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	expectedSeq := 50
	var wg sync.WaitGroup
	for k := 0; k < expectedSeq; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = db.Lock()
			assert.NoError(t, err)
			defer func() {
				err := db.Unlock()
				require.NoError(t, err)
			}()
			err = db.Open()
			require.NoError(t, err)

			_, err = db.Exec(`
				PRAGMA TEMP_STORE = MEMORY;
				PRAGMA JOURNAL_MODE = MEMORY;
				PRAGMA SYNCHRONOUS = OFF;
				PRAGMA LOCKING_MODE = NORMAL;
				UPDATE foo SET seq = (SELECT seq + 1 FROM foo);
			`)
			assert.NoError(t, err)

			err = db.Close()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	err = db.Open()
	require.NoError(t, err)

	var seq int
	row := db.QueryRow(`SELECT seq FROM foo;`)
	row.Scan(&seq)

	assert.Equal(t, expectedSeq, seq)

	err = db.Close()
	require.NoError(t, err)
}

func TestSimultaneousUsage_OneDbByProcess(t *testing.T) {
	// Howto ?
}
