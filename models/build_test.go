/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"log"
	"strconv"
	"testing"

	"github.com/ernestio/service-store/tests"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

const TESTBUILDDB = "test_builds"

// BuildTestSuite : Test suite for migration
type BuildTestSuite struct {
	suite.Suite
}

// SetupTest : sets up test suite
func (suite *BuildTestSuite) SetupTest() {
	err := tests.CreateTestDB(TESTBUILDDB)
	if err != nil {
		log.Fatal(err)
	}

	DB, err = gorm.Open("postgres", "user=postgres dbname="+TESTBUILDDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	//DB.LogMode(true)

	DB.AutoMigrate(Build{})
	DB.Unscoped().Delete(Build{})

	for i := 1; i <= 10; i++ {
		DB.Create(&Build{
			UUID:          "uuid+" + strconv.Itoa(i),
			EnvironmentID: 1,
			UserID:        uint(i),
			Status:        "in_progress",
			Mapping:       map[string]interface{}{},
			Definition:    "yaml",
		})
	}
}

func (suite *BuildTestSuite) TestBuilds() {
	suite.testFindBuilds()
}

func (suite *BuildTestSuite) testFindBuilds() {
	builds, err := FindBuilds(map[string]interface{}{
		"user_id": 1,
	})

	suite.Nil(err)
	suite.Equal(len(builds), 1)
	suite.Equal(builds[0].ID, uint(1))
	suite.Equal(builds[0].UserID, uint(1))
	suite.Equal(builds[0].EnvironmentID, uint(1))
	suite.Equal(builds[0].Definition, "yaml")
}

// TestBuildTestSuite : Test suite for migration
func TestBuildTestSuite(t *testing.T) {
	suite.Run(t, new(BuildTestSuite))
}
