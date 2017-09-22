/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// BuildFind : gets an build
func BuildFind(msg *nats.Msg) {
	var err error
	var q map[string]interface{}
	var builds []models.Build
	var data []byte

	defer response(msg.Reply, &data, &err)

	if len(msg.Data) < 1 {
		msg.Data = []byte(`{}`)
	}

	err = json.Unmarshal(msg.Data, &q)
	if err != nil {
		return
	}

	builds, err = models.FindBuilds(q)
	if err != nil {
		return
	}

	data, err = json.Marshal(builds)
}
