/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"runtime"

	"github.com/ernestio/service-store/handlers"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nats-io/nats"
	"github.com/r3labs/akira"
)

var n akira.Connector
var db *gorm.DB

func startHandler() {
	subscribers := map[string]interface{}{
		"environment.get":             map[string]nats.MsgHandler{"environment-store": handlers.EnvGet},
		"environment.del":             map[string]nats.MsgHandler{"environment-store": handlers.EnvDelete},
		"environment.set":             map[string]nats.MsgHandler{"environment-store": handlers.EnvSet},
		"environment.find":            map[string]nats.MsgHandler{"environment-store": handlers.EnvFind},
		"environment.set.schedule":    map[string]nats.MsgHandler{"environment-store": handlers.SetSchedule},
		"environment.del.schedule":    map[string]nats.MsgHandler{"environment-store": handlers.UnsetSchedule},
		"build.get":                   map[string]nats.MsgHandler{"environment-store": handlers.BuildGet},
		"build.del":                   map[string]nats.MsgHandler{"environment-store": handlers.BuildDelete},
		"build.set":                   map[string]nats.MsgHandler{"environment-store": handlers.BuildSet},
		"build.find":                  map[string]nats.MsgHandler{"environment-store": handlers.BuildFind},
		"build.get.mapping":           map[string]nats.MsgHandler{"environment-store": handlers.BuildGetMapping},
		"build.set.mapping":           map[string]nats.MsgHandler{"environment-store": handlers.BuildSetMapping},
		"build.set.mapping.component": map[string]nats.MsgHandler{"environment-store": handlers.BuildSetComponent},
		"build.del.mapping.component": map[string]nats.MsgHandler{"environment-store": handlers.BuildDeleteComponent},
		"build.set.mapping.change":    map[string]nats.MsgHandler{"environment-store": handlers.BuildSetChange},
		"build.get.definition":        map[string]nats.MsgHandler{"environment-store": handlers.BuildGetDefinition},
		"build.set.definition":        map[string]nats.MsgHandler{"environment-store": handlers.BuildSetDefinition},
		"build.*.done":                map[string]nats.MsgHandler{"environment-store": handlers.BuildComplete},
		"build.*.error":               map[string]nats.MsgHandler{"environment-store": handlers.BuildError},
		"build.set.status":            map[string]nats.MsgHandler{"environment-store": handlers.SetBuildStatus},
	}

	for endpoint, v := range subscribers {
		for store, handler := range v.(map[string]nats.MsgHandler) {
			if _, err := n.QueueSubscribe(endpoint, store, handler); err != nil {
				log.Panic(err)
			}
		}
	}
}

func main() {
	setupNats()
	setupPg("environments")

	err := Migrate(db)
	if err != nil {
		panic(err)
	}

	startHandler()

	runtime.Goexit()
}
