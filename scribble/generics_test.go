package scribble

import (
	"os"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	testStructA struct {
		K0 int
		K0A []int
		Ks string
		KsA [] string
		Kb testStructB
		KbA []testStructB
		Kc testStructC
		KcA []testStructC
		Kd testStructD
		KdA []testStructD
		Km map[testStructC]testStructD
		KmA []map[testStructC]testStructD
	}

	testStructB struct {
		Kc testStructC
		Kd testStructD
	}

	testStructC = string
	testStructD = int
)

var (
	testDbPath = "/tmp/mydb"

	expectedTestStruct0 = testStructA{
		K0: 3,
		K0A: []int{5, 2, 3},
		Ks: "foo",
		KsA: []string{"bar", "baz"},
		Kb: testStructB{Kc: testStructC("pif"), Kd: testStructD(6)},
		KbA: []testStructB{
			testStructB{Kc: testStructC("pef"), Kd: testStructD(7)},
			testStructB{Kc: testStructC("puf"), Kd: testStructD(8)},
		},
		Kc: testStructC("bip"),
		KcA: []testStructC{testStructC("bap"), testStructC("bup")},
		Kd: testStructD(9),
		KdA: []testStructD{testStructD(10), testStructD(11)},
		Km: map[testStructC]testStructD{
			testStructC("k1"): testStructD(12),
			testStructC("k2"): testStructD(13),
		},
		KmA: []map[testStructC]testStructD{
			map[testStructC]testStructD{
				testStructC("k3"): testStructD(14),
				testStructC("k4"): testStructD(15),
			},
			map[testStructC]testStructD{
				testStructC("k5"): testStructD(16),
				testStructC("k6"): testStructD(17),
			},
		},
	}

	expectedTestStructB1 = testStructB{Kc: testStructC("pif"), Kd: testStructD(6)}
	expectedTestStructB2 = testStructB{Kc: testStructC("pef"), Kd: testStructD(7)}
	expectedTestStructB3 = testStructB{Kc: testStructC("puf"), Kd: testStructD(8)}

)

func initDb() *Driver {
	db, err := New("./testScribble", nil)
	if err != nil {
		log.Fatalf("Error initializing DB: %s", err)
	}
	return db
}

func clearDb() {
	os.RemoveAll("./testScribble")
}

func TestReadWrite_String(t *testing.T) {
	db := initDb()
	defer clearDb()

	key := "key"
	expectedValue := "foo"

	var v string
	var err error

	// Read from not existing collection
	v, err = Read[string](db, "strings", key)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Write
	err = Write(db, "strings", key, expectedValue)
	require.NoError(t, err)

	// Read existing
	v, err = Read[string](db, "strings", key)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, v)

	// Read not existing
	v, err = Read[string](db, "strings", "otherKey")
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestReadWrite_Map(t *testing.T) {
	db := initDb()
	defer clearDb()

	key := "key"
	expectedValue := map[string]any{
		"a": "foo",
		"b": "bar",
		"c": []any{
			map[string]any{"k1": "v1", "k2": float64(1)},
			map[string]any{"k1": "v2", "k2": float64(2)},
		},
		"d": map[string]any{"p1": float64(3), "p2": false},
	}

	var v map[string]any
	var err error

	// Read from not existing collection
	v, err = Read[map[string]any](db, "maps", key)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Write
	err = Write(db, "maps", key, expectedValue)
	require.NoError(t, err)

	// Read existing
	v, err = Read[map[string]any](db, "maps", key)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, v)

	// Read not existing
	v, err = Read[map[string]any](db, "maps", "otherKey")
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestReadWrite_Struct(t *testing.T) {
	db := initDb()
	defer clearDb()

	key := "key"

	// Read from not existing collection
	v, err := Read[testStructA](db, "maps", key)
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Write
	err = Write(db, "maps", key, expectedTestStruct0)
	require.NoError(t, err)

	// Read existing
	v, err = Read[testStructA](db, "maps", key)
	require.NoError(t, err)
	assert.Equal(t, expectedTestStruct0, v)

	// Read not existing
	v, err = Read[testStructA](db, "maps", "otherKey")
	require.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestReadAll_Comparable(t *testing.T) {
	db := initDb()
	defer clearDb()

	collection := "myCollection"
	err := Write(db, collection, "k1", "foo")
	require.NoError(t, err)
	err = Write(db, collection, "k2", "bar")
	require.NoError(t, err)
	err = Write(db, collection, "k3", "baz")
	require.NoError(t, err)

	res, err := ReadAll[string](db, collection)
	require.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Contains(t, res, "foo")
	assert.Contains(t, res, "bar")
	assert.Contains(t, res, "baz")

	err = Write(db, collection, "k4", 5)
	require.NoError(t, err)
	_, err = ReadAll[string](db, collection)
	require.Error(t, err)
}

func TestReadAll_Struct(t *testing.T) {
	db := initDb()
	defer clearDb()

	collection := "myCollection"
	err := Write(db, collection, "k1", expectedTestStructB1)
	require.NoError(t, err)
	err = Write(db, collection, "k2", expectedTestStructB2)
	require.NoError(t, err)
	err = Write(db, collection, "k3", expectedTestStructB3)
	require.NoError(t, err)

	res, err := ReadAll[testStructB](db, collection)
	require.NoError(t, err)
	assert.Len(t, res, 3)
	assert.Contains(t, res, expectedTestStructB1)
	assert.Contains(t, res, expectedTestStructB2)
	assert.Contains(t, res, expectedTestStructB3)

	err = Write(db, collection, "k4", 5)
	require.NoError(t, err)
	_, err = ReadAll[testStructB](db, collection)
	require.Error(t, err)
	
}
