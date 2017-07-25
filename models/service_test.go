/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"fmt"
	"log"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

const TESTDB = "test_services"

// ServiceTestSuite : Test suite for migration
type ServiceTestSuite struct {
	suite.Suite
}

// SetupTest : sets up test suite
func (suite *ServiceTestSuite) SetupTest() {
	var err error
	DB, err = gorm.Open("postgres", "user=postgres dbname="+TESTDB+" sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *ServiceTestSuite) TestServiceFind() {
	s := Service{
		Name:    "test-service",
		GroupID: 0,
	}

	services := s.Find()
	fmt.Println(services)
}

// TestServiceTestSuite : Test suite for migration
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
