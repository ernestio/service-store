/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// BuildGet : gets an build
func BuildGet(msg *nats.Msg) {
	var err error
	var q map[string]interface{}
	var build *models.Build
	var data []byte

	defer response(msg.Reply, &data, &err)

	err = json.Unmarshal(msg.Data, &q)
	if err != nil {
		return
	}

	build, err = models.GetBuild(q)
	if err != nil {
		return
	}

	build.Mapping = nil
	build.Definition = ""

	data, err = json.Marshal(build)
}
