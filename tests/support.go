/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package tests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/lib/pq"
)

func CreateTestDB(name string) error {
	db, derr := sql.Open("postgres", "user=postgres sslmode=disable")
	if derr != nil {
		return derr
	}

	_, derr = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(name)))
	if derr != nil {
		return derr
	}

	_, derr = db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(name)))
	if derr != nil {
		return derr
	}

	return nil
}

func CreateMigrationData(dbname string, file string) error {
	db, err := sql.Open("postgres", "user=postgres dbname="+dbname+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	statements := strings.Split(string(data), ";\r")

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, s := range statements {
		_, err := tx.Exec(s)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func LoadJSON(path string, m interface{}) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		panic(err)
	}

}
