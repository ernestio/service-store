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
	"github.com/r3labs/pattern"
)

var n akira.Connector
var db *gorm.DB

func subscriber(subs map[string]nats.MsgHandler, event string) *nats.MsgHandler {
	for k, v := range subs {
		if pattern.Match(event, k) {
			return &v
		}
	}

	return nil
}

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

	_, err := n.Subscribe(">", func(msg *nats.Msg) {
		handler := subscriber(subscribers, msg.Subject)
		if handler != nil {
			(*handler)(msg)
		}
	})

	if err != nil {
		log.Panic(err)
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
