/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// BuildComplete : sets a builds status to complete
func BuildComplete(msg *nats.Msg) {
	var m Message
	var b models.Build

	parts := strings.Split(msg.Subject, ".")

	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		log.Println("could not load completion event: " + err.Error())
	}

	err = b.SetStatus(m.ID, "done")
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
		return
	}

	if parts[1] == "delete" {
		e, err := models.GetEnvironment(map[string]interface{}{"id": b.EnvironmentID})
		if err != nil {
			log.Println("could not get service from service complete message: " + err.Error())
		}

		err = e.Delete()
		if err != nil {
			log.Println("could not get delete the service: " + err.Error())
		}

		DetatchPolicies(e.Name)
	}
}
