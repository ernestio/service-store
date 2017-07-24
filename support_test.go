/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

func setupTestSuite() {
	db.Unscoped().Delete(Entity{})
}

func createEntities(n int) {
	i := 0
	for i < n {
		x := strconv.Itoa(i)
		db.Create(&Entity{
			Name:         "Test" + x,
			UUID:         "random_string" + x,
			GroupID:      1,
			DatacenterID: 1,
			Type:         "type",
			Version:      time.Now(),
			Status:       "in_progress",
			Options:      "options",
			Definition:   "definition",
			Mapping:      `{"id": "random_string` + x + `", "action": "service.create", "components":[{"_component_id": "network::test-1", "_state": "running"}, {"_component_id": "network::test-2", "_state": "running"}], "changes":[{"_component_id": "network::test-3", "_state": "waiting"}, {"_component_id": "network::test-4", "_state": "waiting"}]}`,
		})
		i++
	}
}

func createTestDB(name string) error {
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

func createTestData(db *sql.DB, file string) error {
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
