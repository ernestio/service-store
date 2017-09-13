/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// BuildDelete : gets an build
func BuildDelete(msg *nats.Msg) {
	var err error
	var build *models.Build
	var data []byte

	defer response(msg.Reply, &data, &err)

	err = json.Unmarshal(msg.Data, build)
	if err != nil {
		return
	}

	if build.ID == 0 {
		err = build.Create()
		return
	}

	err = build.Delete()

	if err != nil {
		return
	}

	data = []byte(`{"status": "success"}`)
}
