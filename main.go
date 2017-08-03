/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"
	"runtime"

	"github.com/jinzhu/gorm"
	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var n *nats.Conn
var db *gorm.DB

var handler natsdb.Handler

func startHandler() {
	handler = natsdb.Handler{
		NotFoundErrorMessage:   natsdb.NotFound.Encoded(),
		UnexpectedErrorMessage: natsdb.Unexpected.Encoded(),
		DeletedMessage:         []byte(`{"status":"deleted"}`),
		Nats:                   n,
		NewModel: func() natsdb.Model {
			return &ServiceView{}
		},
	}

	if _, err := n.QueueSubscribe("service.get", "service-store", handler.Get); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.del", "service-store", handler.Del); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.set", "service-store", handler.Set); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.find", "service-store", handler.Find); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.get.mapping", "service-store", GetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.set.mapping", "service-store", SetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.set.mapping.component", "service-store", SetComponent); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.del.mapping.component", "service-store", DeleteComponent); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.set.mapping.change", "service-store", SetChange); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.get.definition", "service-store", GetDefinition); err != nil {
		log.Panic(err)
	}
	if _, err := n.QueueSubscribe("service.set.definition", "service-store", SetDefinition); err != nil {
		log.Panic(err)
	}
}

func main() {
	setupNats()
	setupPg("services")
	startHandler()

	Migrate(db)

	runtime.Goexit()
}
