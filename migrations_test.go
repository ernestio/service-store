/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"testing"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/suite"
)

const TESTDB = "services_test"

// MigrationTestSuite : Test suite for migration
type MigrationTestSuite struct {
	suite.Suite
	DB *sql.DB
}

// SetupTest : sets up test suite
func (suite *MigrationTestSuite) SetupTest() {
	err := createTestDB(TESTDB)
	if err != nil {
		log.Fatal(err)
	}

	suite.DB, err = sql.Open("postgres", "user=postgres dbname="+TESTDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = createTestData(suite.DB, "tests/sql/services_test.sql")
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *MigrationTestSuite) TestMigration() {
	err := Migrate(suite.DB)

	suite.Nil(err)

	rows, err := suite.DB.Query("SELECT COUNT(*) AS count FROM services;")
	suite.Nil(err)

	suite.Nil(err)
	//suite.Equal(count, 41)
}

// TestMigrationTestSuite : Test suite for migration
func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
