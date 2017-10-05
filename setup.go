/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"os"

	ecc "github.com/ernestio/ernest-config-client"
	"github.com/ernestio/service-store/handlers"
	"github.com/ernestio/service-store/models"
)

var c *ecc.Config

func setupNats() {
	c = ecc.NewConfig(os.Getenv("NATS_URI"))
	n = c.Nats()
	handlers.NC = n
}

func setupPg(dbname string) {
	db = c.Postgres(dbname)
	models.DB = db
}
