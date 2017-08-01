/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/ernestio/service-store/models"
	"github.com/ernestio/service-store/tests"
	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats"
	. "github.com/smartystreets/goconvey/convey"
)

func CreateTestData(db *gorm.DB, count int) {
	for i := 1; i <= count; i++ {
		db.Create(&models.Service{
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

	for i := 1; i <= count; i++ {
		db.Create(&models.Build{
			UUID:      "uuid-" + strconv.Itoa(i),
			ServiceID: uint(i),
			UserID:    uint(i),
			Status:    "in_progress",
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

func TestHandler(t *testing.T) {
	tests.CreateTestDB("test_handlers")

	setupNats()

	_, _ = n.Subscribe("config.get.postgres", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"names":["services"],"password":"","url":"postgres://postgres@127.0.0.1","user":""}`))
	})

	setupPg("test_handlers")
	db.AutoMigrate(models.Service{}, models.Build{})

	startHandler()

	db.Unscoped().Delete(models.Service{}, models.Build{})
	CreateTestData(db, 20)

	Convey("Scenario: getting a service", t, func() {
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.get", []byte(`{"id":"uuid-9999"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.NotFoundErrorMessage))
			So(err, ShouldBeNil)
		})

		Convey("Given the service exists on the database", func() {
			id := "uuid-1"

			msg, err := n.Request("service.get", []byte(`{"id":"`+id+`"}`), time.Second)
			output := ServiceView{}
			_ = json.Unmarshal(msg.Data, &output)

			So(output.UUID, ShouldEqual, "uuid-1")
			So(output.Name, ShouldEqual, "Test1")
			So(err, ShouldBeNil)
		})

		Convey("Given the service exists on the database and searching by name", func() {
			name := "Test3"

			msg, err := n.Request("service.get", []byte(`{"name":"`+name+`"}`), time.Second)
			output := ServiceView{}
			_ = json.Unmarshal(msg.Data, &output)

			So(output.UUID, ShouldEqual, "uuid-3")
			So(output.GroupID, ShouldEqual, 1)
			So(output.DatacenterID, ShouldEqual, 1)
			So(output.Name, ShouldEqual, "Test3")
			So(output.Version, ShouldNotBeNil)
			So(output.Status, ShouldEqual, "in_progress")
			So(output.Options["sync"], ShouldBeTrue)
			So(err, ShouldBeNil)
		})
	})

	Convey("Scenario: find services", t, func() {
		Convey("Given services exist on the database", func() {
			Convey("Then I should get a list of services", func() {
				msg, err := n.Request("service.find", []byte(`{"group_id":1}`), time.Second)
				So(err, ShouldBeNil)

				list := []ServiceView{}
				_ = json.Unmarshal(msg.Data, &list)
				So(len(list), ShouldEqual, 20)
				So(list[0].Name, ShouldEqual, "Test20")
				So(list[0].UUID, ShouldEqual, "uuid-20")
				So(list[0].GroupID, ShouldEqual, 1)
				So(list[0].UserID, ShouldEqual, 20)
				So(list[0].Status, ShouldEqual, "in_progress")
				So(list[19].Name, ShouldEqual, "Test1")
				So(list[19].UUID, ShouldEqual, "uuid-1")
				So(list[19].GroupID, ShouldEqual, 1)
				So(list[19].UserID, ShouldEqual, 1)
				So(list[19].Status, ShouldEqual, "in_progress")
			})
		})
	})

	Convey("Scenario: deleting a service", t, func() {
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.del", []byte(`{"id":"32"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.NotFoundErrorMessage))
			So(err, ShouldBeNil)
		})

		Convey("Given the service exists on the database", func() {
			id := "uuid-8"

			msg, err := n.Request("service.del", []byte(`{"id":"`+id+`"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.DeletedMessage))
			So(err, ShouldBeNil)
		})
	})

	Convey("Scenario: service set", t, func() {
		Convey("Given we don't provide any id as part of the body", func() {
			Convey("Then it should return the created record and it should be stored on DB", func() {
				msg, err := n.Request("service.set", []byte(`{"name":"test-1"}`), time.Second)
				output := ServiceView{}
				output.LoadFromInput(msg.Data)
				So(output.ID, ShouldNotEqual, 0)
				So(output.UUID, ShouldNotBeNil)
				So(output.Name, ShouldEqual, "test-1")
				So(err, ShouldBeNil)
			})
		})

		Convey("Given we provide an unexisting id", func() {
			Convey("Then it should store the service", func() {
				msg, err := n.Request("service.set", []byte(`{"id": "unexisting", "name":"test-2"}`), time.Second)
				output := ServiceView{}
				output.LoadFromInput(msg.Data)
				So(output.UUID, ShouldEqual, "unexisting")
				So(output.Name, ShouldEqual, "test-2")
				So(err, ShouldBeNil)
			})
		})

		Convey("Given we provide an existing id", func() {
			Convey("When I update an existing entity", func() {
				id := "uuid-4"

				msg, err := n.Request("service.set", []byte(`{"id": "`+id+`", "options":{"sync":false}}`), time.Second)
				So(err, ShouldBeNil)
				output := ServiceView{}
				output.LoadFromInput(msg.Data)
				Convey("Then we should receive an updated entity", func() {
					So(output.UUID, ShouldEqual, id)
				})
				Convey("And non provided fields should not be updated", func() {
					So(output.Options["sync"], ShouldBeTrue)
				})
			})
		})
	})

	Convey("Scenario: getting setting a service mapping", t, func() {
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.get.mapping", []byte(`{"id":"32"}`), time.Second)
			So(string(msg.Data), ShouldEqual, `{"error": "bad request"}`)
			So(err, ShouldBeNil)
		})

		Convey("And the service exists on the database", func() {
			id := "uuid-3"
			Convey("Then calling service.get.mapping should return the valid mapping", func() {
				msg, err := n.Request("service.get.mapping", []byte(`{"id":"`+id+`"}`), time.Second)
				So(string(msg.Data), ShouldEqual, `{"action":"service.create","changes":[{"_component_id":"network::test-3","_state":"waiting"},{"_component_id":"network::test-4","_state":"waiting"}],"components":[{"_component_id":"network::test-1","_state":"running"},{"_component_id":"network::test-2","_state":"running"}],"id":"uuid-3"}`)
				So(err, ShouldBeNil)
			})
			Convey("And calling service.set.mapping should update mapping", func() {
				msg, err := n.Request("service.set.mapping", []byte(`{"id":"`+id+`","mapping":{"updated":"content"}}`), time.Second)
				So(string(msg.Data), ShouldEqual, `{"status": "success"}`)
				So(err, ShouldBeNil)
				msg, err = n.Request("service.get.mapping", []byte(`{"id":"`+id+`"}`), time.Second)
				So(string(msg.Data), ShouldEqual, `{"updated":"content"}`)
				So(err, ShouldBeNil)
			})

		})
	})

}
