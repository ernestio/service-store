/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"testing"

	"github.com/ernestio/service-store/models"
	"github.com/ernestio/service-store/tests"
	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/suite"
)

const TESTMIGRATIONDB = "test_migrations"

// MigrationTestSuite : Test suite for migration
type MigrationTestSuite struct {
	suite.Suite
	DB       *gorm.DB
	services map[string]models.Service
	builds   map[string]models.Build
}

// SetupTest : sets up test suite
func (suite *MigrationTestSuite) SetupTest() {
	err := tests.CreateTestDB(TESTMIGRATIONDB)
	if err != nil {
		log.Fatal(err)
	}

	err = tests.CreateMigrationData(TESTMIGRATIONDB, "tests/sql/services_test.sql")
	if err != nil {
		log.Fatal(err)
	}

	suite.DB, err = gorm.Open("postgres", "user=postgres dbname="+TESTMIGRATIONDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	tests.LoadJSON("./tests/json/valid-services.json", &suite.services)
	tests.LoadJSON("./tests/json/valid-builds.json", &suite.builds)
}

func (suite *MigrationTestSuite) TestMigration() {
	var builds []models.Build
	var services []models.Service

	err := Migrate(suite.DB)
	suite.Nil(err)

	suite.DB.Table("builds").Find(&builds)
	suite.Equal(len(builds), 41)

	for _, b := range builds {
		vb := suite.builds[b.UUID]
		suite.NotNil(vb)
		suite.Equal(b.ServiceID, vb.ServiceID)
		suite.Equal(b.UserID, vb.UserID)
		suite.Equal(b.Type, vb.Type)
		suite.Equal(b.Status, vb.Status)
		suite.NotEqual(b.Definition, "")

		action, ok := b.Mapping["action"].(string)
		suite.True(ok)
		suite.Equal(action, "service.create")
	}

	suite.DB.Table("services").Find(&services)
	suite.Equal(len(services), 21)
	for _, s := range services {
		vs := suite.services[s.Name]
		suite.NotNil(vs)
		suite.Equal(s.ID, vs.ID)
		suite.Equal(s.GroupID, vs.GroupID)
		suite.Equal(s.DatacenterID, vs.DatacenterID)
		suite.Equal(s.Status, vs.Status)
	}
}

// TestMigrationTestSuite : Test suite for migration
func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
