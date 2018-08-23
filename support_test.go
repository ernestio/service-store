/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"strconv"

	ecc "github.com/ernestio/ernest-config-client"
	"github.com/ernestio/service-store/handlers"
	"github.com/ernestio/service-store/models"
	"github.com/ernestio/service-store/tests"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/go-nats"
	"github.com/r3labs/akira"
)

func CreateTestData(db *gorm.DB, count int) {
	for i := 1; i <= count; i++ {
		db.Create(&models.Environment{
			Name:   "Test" + strconv.Itoa(i),
			Status: "done",
			Options: map[string]interface{}{
				"sync":          true,
				"sync_type":     "hard",
				"sync_interval": 5,
			},
			Credentials: map[string]interface{}{
				"username": "test",
				"password": "test",
			},
		})
	}

	for i := 1; i <= count; i++ {
		db.Create(&models.Build{
			UUID:          "uuid-" + strconv.Itoa(i),
			EnvironmentID: uint(i),
			UserID:        uint(i),
			Status:        "done",
			Type:          "apply",
			Mapping: map[string]interface{}{
				"id":     "uuid-" + strconv.Itoa(i),
				"action": "service.create",
				"components": []map[string]interface{}{
					{
						"_component_id": "network::test-1",
						"_state":        "running",
					},
					{
						"_component_id": "network::test-2",
						"_state":        "running",
					},
				},
				"changes": []map[string]interface{}{
					{
						"_component_id": "network::test-3",
						"_state":        "waiting",
					},
					{
						"_component_id": "network::test-4",
						"_state":        "waiting",
					},
				},
			},
			Definition: "yaml",
		})
	}
}

func setupTestSuite(database string) {
	n = akira.NewFakeConnector()
	handlers.NC = n

	c = &ecc.Config{}
	c.SetConnector(n)

	_, _ = n.Subscribe("config.get.postgres", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"names":["services"],"password":"","url":"postgres://postgres@127.0.0.1","user":""}`))
	})

	_ = tests.CreateTestDB(database)
	setupPg(database)
	db.AutoMigrate(models.Environment{}, models.Build{})

	startHandler()
}
