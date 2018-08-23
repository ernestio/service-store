/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/go-nats"
)

// BuildSetValidation : Validation field setter
func BuildSetValidation(msg *nats.Msg) {
	var err error
	var m *Message
	var b *models.Build

	defer response(msg.Reply, nil, &err)

	err = json.Unmarshal(msg.Data, &m)
	if err != nil {
		return
	}

	b, err = models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		return
	}

	b.Validation = m.Validation

	err = b.Update()
}
