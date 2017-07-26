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

const TESTSERVICEDB = "test_services"

// ServiceTestSuite : Test suite for migration
type ServiceTestSuite struct {
	suite.Suite
}

// SetupTest : sets up test suite
func (suite *ServiceTestSuite) SetupTest() {
	err := tests.CreateTestDB(TESTSERVICEDB)
	if err != nil {
		log.Fatal(err)
	}

	DB, err = gorm.Open("postgres", "user=postgres dbname="+TESTSERVICEDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	//DB.LogMode(true)

	DB.AutoMigrate(Service{})
	DB.Unscoped().Delete(Service{})

	for i := 1; i <= 10; i++ {
		DB.Create(&Service{
			Name:         "Test" + strconv.Itoa(i),
			GroupID:      1,
			DatacenterID: 1,
			Status:       "in_progress",
			Options: map[string]interface{}{
				"sync":          true,
				"sync_type":     "hard",
				"sync_interval": 5,
			},
			Credentials: map[string]interface{}{},
		})
	}
}

func (suite *ServiceTestSuite) TestServices() {
	suite.testFindServices()
}

func (suite *ServiceTestSuite) testFindServices() {
	services := FindServices(map[string]interface{}{
		"name":     "Test1",
		"group_id": 1,
	})

	suite.Equal(len(services), 1)
	suite.Equal(services[0].ID, uint(1))
	suite.Equal(services[0].GroupID, uint(1))
	suite.Equal(services[0].DatacenterID, uint(1))
	suite.Equal(services[0].Name, "Test1")
	suite.Equal(services[0].Status, "in_progress")
	suite.Equal(services[0].Options["sync"], true)
}

// TestServiceTestSuite : Test suite for migration
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
