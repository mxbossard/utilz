package scribble

import (
	"encoding/json"
	"errors"
	"os"
)

func Read[T any](db *Driver, collection, resource string) (T, error) {
	var v T
	err := db.Read(collection, resource, &v)
	return v, err
}

func Write[T any](db *Driver, collection, resource string, v T) error {
	err := db.Write(collection, resource, v)
	return err
}

func ReadAll[T any](db *Driver, collection string) ([]T, error) {
	var items []T
	records, err := db.ReadAll(collection)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		var i T
		if err := json.Unmarshal(r, &i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func ReadAllOrEmpty[T any](db *Driver, collection string) ([]T, error) {
	res, err := ReadAll[T](db, collection)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return res, err
}