/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/go-nats"
	"github.com/r3labs/graph"
)

// BuildSetComponent : Mapping component setter
func BuildSetComponent(msg *nats.Msg) {
	var err error
	var b models.Build
	var c graph.GenericComponent

	defer response(msg.Reply, nil, &err)

	err = json.Unmarshal(msg.Data, &c)
	if err != nil {
		return
	}

	err = b.SetComponent(&c)
}
