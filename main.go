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
var err error
var handler natsdb.Handler

func startHandler() {
	handler = natsdb.Handler{
		NotFoundErrorMessage:   natsdb.NotFound.Encoded(),
		UnexpectedErrorMessage: natsdb.Unexpected.Encoded(),
		DeletedMessage:         []byte(`{"status":"deleted"}`),
		Nats:                   n,
		NewModel: func() natsdb.Model {
			return &Entity{}
		},
	}

	if _, err := n.Subscribe("service.get", handler.Get); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.del", handler.Del); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.set", handler.Set); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.find", handler.Find); err != nil {
		log.Panic(err)
	}

	if _, err := n.Subscribe("service.get.mapping", GetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.set.mapping", SetMapping); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.get.definition", GetDefinition); err != nil {
		log.Panic(err)
	}
	if _, err := n.Subscribe("service.set.definition", SetDefinition); err != nil {
		log.Panic(err)
	}
}

func main() {
	setupNats()
	setupPg()
	startHandler()

	runtime.Goexit()
}
