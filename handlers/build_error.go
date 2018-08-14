/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"
	"log"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/go-nats"
)

// BuildError : sets a builds status to errored
func BuildError(msg *nats.Msg) {
	var m Message
	var b models.Build

	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}

	err = b.SetStatus(m.ID, "errored")
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}
}
