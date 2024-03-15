package db

import "fmt"

var database Idb

func GetDb() (Idb, error) {
	if database == nil {
		return nil, fmt.Errorf("no database has been set")
	}
	return database, nil
}

func SetDb(db Idb) {
	database = db
}
