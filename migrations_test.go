/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"testing"

	"github.com/jinzhu/gorm"
	//	_ "github.com/lib/pq"

	"github.com/stretchr/testify/suite"
)

const TESTDB = "test_services"

// MigrationTestSuite : Test suite for migration
type MigrationTestSuite struct {
	suite.Suite
	DB *gorm.DB
}

// SetupTest : sets up test suite
func (suite *MigrationTestSuite) SetupTest() {
	err := createTestDB(TESTDB)
	if err != nil {
		log.Fatal(err)
	}

	err = createTestData(TESTDB, "tests/sql/services_test.sql")
	if err != nil {
		log.Fatal(err)
	}

	suite.DB, err = gorm.Open("postgres", "user=postgres dbname="+TESTDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *MigrationTestSuite) TestMigration() {
	scount := 0
	bcount := 0

	err := Migrate(suite.DB)
	suite.Nil(err)

	suite.DB.Table("builds").Count(&bcount)
	suite.Equal(bcount, 41)

	suite.DB.Table("services").Count(&scount)
	suite.Equal(scount, 21)
}

// TestMigrationTestSuite : Test suite for migration
func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
