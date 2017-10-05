/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// EnvGet : gets an environment
func EnvGet(msg *nats.Msg) {
	var err error
	var q map[string]interface{}
	var env *models.Environment
	var data []byte

	defer response(msg.Reply, &data, &err)

	err = json.Unmarshal(msg.Data, &q)
	if err != nil {
		return
	}

	env, err = models.GetEnvironment(q)
	if err != nil {
		return
	}

	data, err = json.Marshal(env)
}
