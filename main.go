/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
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
	subscribers := map[string]nats.MsgHandler{
		"environment.get":             handlers.EnvGet,
		"environment.del":             handlers.EnvDelete,
		"environment.set":             handlers.EnvSet,
		"environment.find":            handlers.EnvFind,
		"environment.set.schedule":    handlers.SetSchedule,
		"environment.del.schedule":    handlers.UnsetSchedule,
		"build.get":                   handlers.BuildGet,
		"build.del":                   handlers.BuildDelete,
		"build.set":                   handlers.BuildSet,
		"build.find":                  handlers.BuildFind,
		"build.get.mapping":           handlers.BuildGetMapping,
		"build.set.mapping":           handlers.BuildSetMapping,
		"build.set.mapping.component": handlers.BuildSetComponent,
		"build.del.mapping.component": handlers.BuildDeleteComponent,
		"build.set.mapping.change":    handlers.BuildSetChange,
		"build.get.definition":        handlers.BuildGetDefinition,
		"build.set.definition":        handlers.BuildSetDefinition,
		"build.*.done":                handlers.BuildComplete,
		"build.*.error":               handlers.BuildError,
		"build.set.status":            handlers.SetBuildStatus,
	}

	n.Subscribe(">", func(msg *nats.Msg) {
		handler, ok := subscribers[msg.Subject]
		if ok {
			handler(msg)
		}
	})
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
