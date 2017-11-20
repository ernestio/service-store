/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"
	"log"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// SetSchedule : sets the a schedule for a specific environment
func SetSchedule(msg *nats.Msg) {
	var err error
	var resp []byte
	var env *models.Environment
	var req map[string]interface{}

	defer response(msg.Reply, &resp, &err)

	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}

	if _, ok := req["id"]; !ok {
		resp = []byte(`{"status": "error"}`)
		log.Println("[ ERROR ] a valid id must be provided")
		return
	}

	q := map[string]interface{}{"name": req["name"]}
	env, err = models.GetEnvironment(q)
	if err != nil {
		log.Println("[ ERROR ] retrieving environment info when setting a schedule")
		return
	}

	env.SetSchedule(req["id"].(string), req)
	resp = []byte(`{"status": "success"}`)
}
