/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetHandler(t *testing.T) {
	setupNats()
	_, _ = n.Subscribe("config.get.postgres", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"names":["users","services","services","services"],"password":"","url":"postgres://postgres@127.0.0.1","user":""}`))
	})
	_, _ = n.Subscribe("definition.map.service", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"my":"definition"}`))
	})
	setupPg()
	startHandler()

	Convey("Scenario: getting a service", t, func() {
		setupTestSuite()
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.get", []byte(`{"id":"32"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.NotFoundErrorMessage))
			So(err, ShouldEqual, nil)
		})

		Convey("Given the service exists on the database", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)
			msg, err := n.Request("service.get", []byte(`{"id":"`+id+`"}`), time.Second)
			output := Entity{}
			_ = json.Unmarshal(msg.Data, &output)
			So(output.UUID, ShouldEqual, e.UUID)
			So(output.Name, ShouldEqual, e.Name)
			So(output.Type, ShouldEqual, e.Type)
			So(err, ShouldEqual, nil)
		})

		Convey("Given the service exists on the database and searching by name", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)

			msg, err := n.Request("service.get", []byte(`{"name":"`+e.Name+`"}`), time.Second)
			output := Entity{}
			_ = json.Unmarshal(msg.Data, &output)
			So(output.UUID, ShouldEqual, e.UUID)
			So(output.GroupID, ShouldEqual, e.GroupID)
			So(output.DatacenterID, ShouldEqual, e.DatacenterID)
			So(output.Name, ShouldEqual, e.Name)
			So(output.Type, ShouldEqual, e.Type)
			So(output.Version, ShouldNotBeNil)
			So(output.Status, ShouldEqual, e.Status)
			So(output.Options, ShouldEqual, e.Options)
			So(output.Definition, ShouldEqual, e.Definition)
			So(output.Mapping, ShouldEqual, e.Mapping)
			So(err, ShouldEqual, nil)
		})
	})

	Convey("Scenario: deleting a service", t, func() {
		setupTestSuite()
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.del", []byte(`{"id":"32"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.NotFoundErrorMessage))
			So(err, ShouldEqual, nil)
		})

		Convey("Given the service exists on the database", func() {
			createEntities(1)
			last := Entity{}
			db.First(&last)
			id := fmt.Sprint(last.UUID)

			msg, err := n.Request("service.del", []byte(`{"id":"`+id+`"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.DeletedMessage))
			So(err, ShouldEqual, nil)

			deleted := Entity{}
			db.Where("uuid = ?", id).First(&deleted)
			So(deleted.UUID, ShouldEqual, "")
		})
	})

	Convey("Scenario: service set", t, func() {
		setupTestSuite()
		Convey("Given we don't provide any id as part of the body", func() {
			Convey("Then it should return the created record and it should be stored on DB", func() {
				msg, err := n.Request("service.set", []byte(`{"name":"fred"}`), time.Second)
				output := Entity{}
				output.LoadFromInput(msg.Data)
				So(output.UUID, ShouldNotEqual, nil)
				So(output.Name, ShouldEqual, "fred")
				So(err, ShouldEqual, nil)

				stored := Entity{}
				db.Where("uuid = ?", output.UUID).First(&stored)
				So(stored.Name, ShouldEqual, "fred")
			})
		})

		Convey("Given we provide an unexisting id", func() {
			Convey("Then it should store the service", func() {
				msg, err := n.Request("service.set", []byte(`{"id": "unexisting", "name":"fred"}`), time.Second)
				output := Entity{}
				output.LoadFromInput(msg.Data)
				So(output.UUID, ShouldEqual, "unexisting")
				So(output.Name, ShouldEqual, "fred")
				So(err, ShouldEqual, nil)
			})
		})

		Convey("Given we provide an existing id", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)
			Convey("When I update an existing entity", func() {
				msg, err := n.Request("service.set", []byte(`{"id": "`+id+`", "name":"fred"}`), time.Second)
				output := Entity{}
				output.LoadFromInput(msg.Data)
				stored := Entity{}
				db.Where("uuid = ?", output.UUID).First(&stored)
				Convey("Then we should receive an updated entity", func() {
					So(output.UUID, ShouldNotEqual, nil)
					So(output.Name, ShouldEqual, "fred")
					So(err, ShouldEqual, nil)

					So(stored.Name, ShouldEqual, "fred")
				})
				Convey("And non provided fields should not be updated", func() {
					So(stored.Status, ShouldEqual, e.Status)
					So(stored.UUID, ShouldEqual, e.UUID)
					So(stored.GroupID, ShouldEqual, e.GroupID)
					So(stored.DatacenterID, ShouldEqual, e.DatacenterID)
					So(stored.Type, ShouldEqual, e.Type)
					So(stored.Version, ShouldNotBeNil)
					So(stored.Options, ShouldEqual, e.Options)
					So(stored.Definition, ShouldEqual, e.Definition)
					So(stored.Mapping, ShouldEqual, e.Mapping)

					So(output.Status, ShouldEqual, e.Status)
					So(output.UUID, ShouldEqual, e.UUID)
					So(output.GroupID, ShouldEqual, e.GroupID)
					So(output.DatacenterID, ShouldEqual, e.DatacenterID)
					So(output.Type, ShouldEqual, e.Type)
					So(output.Version, ShouldNotBeNil)
					So(output.Options, ShouldEqual, e.Options)
					So(output.Definition, ShouldEqual, e.Definition)
					So(output.Mapping, ShouldEqual, e.Mapping)
				})
			})
		})
	})

	Convey("Scenario: find services", t, func() {
		setupTestSuite()
		Convey("Given services exist on the database", func() {
			createEntities(20)
			Convey("Then I should get a list of services", func() {
				msg, _ := n.Request("service.find", []byte(`{"group_id":1}`), time.Second)
				list := []Entity{}
				_ = json.Unmarshal(msg.Data, &list)
				So(len(list), ShouldEqual, 20)
				s := list[0]
				So(s.Name, ShouldEqual, "Test19")
				So(s.Endpoint, ShouldEqual, "1.1.1.1")

				stored := Entity{}
				db.Where("uuid = ?", s.UUID).First(&stored)
				So(stored.Endpoint, ShouldEqual, "")
			})
		})
	})

	Convey("Scenario: getting setting a service mapping", t, func() {
		setupTestSuite()
		Convey("Given the service does not exist on the database", func() {
			msg, err := n.Request("service.get.mapping", []byte(`{"id":"32"}`), time.Second)
			So(string(msg.Data), ShouldEqual, string(handler.NotFoundErrorMessage))
			So(err, ShouldEqual, nil)
		})

		Convey("And the service exists on the database", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)
			Convey("Then calling service.get.mapping should return the valid mapping", func() {
				msg, err := n.Request("service.get.mapping", []byte(`{"id":"`+id+`"}`), time.Second)
				So(string(msg.Data), ShouldEqual, string(e.Mapping))
				So(err, ShouldEqual, nil)
			})
			Convey("And calling service.set.mapping should update mapping", func() {
				msg, err := n.Request("service.set.mapping", []byte(`{"id":"`+id+`","mapping":"{\"updated\":\"content\"}"}`), time.Second)
				So(string(msg.Data), ShouldEqual, `"success"`)
				So(err, ShouldEqual, nil)
				msg, err = n.Request("service.get.mapping", []byte(`{"id":"`+id+`"}`), time.Second)
				So(string(msg.Data), ShouldEqual, `{"updated":"content"}`)
				So(err, ShouldEqual, nil)
			})

		})
	})

}
