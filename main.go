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
)

var n *nats.Conn
var db *gorm.DB

func startHandler() {
	// Environments
	if _, err := n.QueueSubscribe("environment.get", "environment-store", handlers.EnvGet); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("environment.del", "environment-store", handlers.EnvDelete); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("environment.set", "environment-store", handlers.EnvSet); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("environment.find", "environment-store", handlers.EnvFind); err != nil {
		log.Panic(err)
	}

	// Builds
	if _, err := n.QueueSubscribe("build.get", "environment-store", handlers.BuildGet); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.del", "environment-store", handlers.BuildDelete); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set", "environment-store", handlers.BuildSet); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.find", "environment-store", handlers.BuildFind); err != nil {
		log.Panic(err)
	}

	if _, err := n.QueueSubscribe("build.get.mapping", "environment-store", handlers.BuildGetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set.mapping", "environment-store", handlers.BuildSetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set.mapping.component", "environment-store", handlers.BuildSetComponent); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.del.mapping.component", "environment-store", handlers.BuildDeleteComponent); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set.mapping.change", "environment-store", handlers.BuildSetChange); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.get.definition", "environment-store", handlers.BuildGetDefinition); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set.definition", "environment-store", handlers.BuildSetDefinition); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.*.done", "environment-store", handlers.BuildComplete); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.*.error", "environment-store", handlers.BuildError); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("build.set.status", "environment-store", SetBuildStatus); err != nil {
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
