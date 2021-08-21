package database

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Database struct {
	DatabasePath string
	Mu           sync.Mutex

	Querys   []string
	Searched []string
	Posted   []string
	Titles   []Title
}

func (db *Database) GetPostsCount() int {
	return len(db.Titles)
}

func (db *Database) Transaction(f func()) {
	db.Mu.Lock()
	defer db.Mu.Unlock()
	f()
}

func (db *Database) ReadDB() error {
	db.Mu.Lock()
	defer db.Mu.Unlock()

	file, err := os.Open(db.DatabasePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var appData AppData
	err = json.Unmarshal(data, &appData)
	if err != nil {
		return err
	}

	db.Querys = appData.Querys
	db.Searched = appData.Searched
	db.Posted = appData.Posted
	return nil
}

func (db *Database) UpdateDB() error {
	db.Mu.Lock()
	defer db.Mu.Unlock()

	appData := AppData{
		Querys:   db.Querys,
		Searched: db.Searched,
		Posted:   db.Posted,
		Titles:   db.Titles,
	}

	data, err := json.Marshal(appData)
	if err != nil {
		return err
	}

	file, err := os.Create(db.DatabasePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func NewDatabase(databasePath string) *Database {
	return &Database{
		DatabasePath: databasePath,
	}
}
