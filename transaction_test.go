/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetComponentHandler(t *testing.T) {
	setupNats()
	_, _ = n.Subscribe("config.get.postgres", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"names":["users","services","services","services"],"password":"","url":"postgres://postgres@127.0.0.1","user":""}`))
	})
	_, _ = n.Subscribe("definition.map.service", func(msg *nats.Msg) {
		_ = n.Publish(msg.Reply, []byte(`{"my":"definition"}`))
	})
	setupPg()
	startHandler()

	Convey("Scenario: setting multiple components on a service concurrently", t, func() {

		setupTestSuite()
		Convey("When receiving two events that update the same service mapping", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)

			_ = n.Publish("service.set.mapping.component", []byte(`{"_component_id":"network::test-1", "service":"`+id+`", "_state": "completed"}`))
			_, err := n.Request("service.set.mapping.component", []byte(`{"_component_id":"network::test-2", "service":"`+id+`", "_state": "completed"}`), time.Second)

			Convey("It should update both the components", func() {
				var m Mapping
				So(err, ShouldBeNil)
				db.First(&e)
				lerr := m.Load([]byte(e.Mapping))
				So(lerr, ShouldBeNil)
				c1, err := m.GetComponent("network::test-1")
				So(err, ShouldBeNil)
				So((*c1)["_state"].(string), ShouldEqual, "completed")
				c2, err := m.GetComponent("network::test-2")
				So(err, ShouldBeNil)
				So((*c2)["_state"].(string), ShouldEqual, "completed")
			})
		})

		Convey("When receiving an event that deletes a component", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)

			_, err := n.Request("service.del.mapping.component", []byte(`{"_component_id":"network::test-2", "service":"`+id+`", "_state": "completed"}`), time.Second)

			Convey("It should remove it from the mapping", func() {
				var m Mapping
				So(err, ShouldBeNil)
				db.First(&e)
				lerr := m.Load([]byte(e.Mapping))
				So(lerr, ShouldBeNil)
				c1, err := m.GetComponent("network::test-1")
				So(err, ShouldBeNil)
				So((*c1)["_state"].(string), ShouldEqual, "completed")
				_, err = m.GetComponent("network::test-2")
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Scenario: setting multiple changes on a service concurrently", t, func() {

		setupTestSuite()
		Convey("When receiving two events that update the same service mapping", func() {
			createEntities(1)
			e := Entity{}
			db.First(&e)
			id := fmt.Sprint(e.UUID)

			_ = n.Publish("service.set.mapping.change", []byte(`{"_component_id":"network::test-3", "service":"`+id+`", "_state": "completed"}`))
			_, err := n.Request("service.set.mapping.change", []byte(`{"_component_id":"network::test-4", "service":"`+id+`", "_state": "completed"}`), time.Second)

			Convey("It should update both the components", func() {
				So(err, ShouldBeNil)

				var m Mapping
				db.First(&e)

				lerr := m.Load([]byte(e.Mapping))
				So(lerr, ShouldBeNil)

				c1, err := m.GetChange("network::test-3")
				So(err, ShouldBeNil)

				So((*c1)["_state"].(string), ShouldEqual, "completed")
				c2, err := m.GetChange("network::test-4")

				So(err, ShouldBeNil)
				So((*c2)["_state"].(string), ShouldEqual, "completed")
			})
		})

	})

}
