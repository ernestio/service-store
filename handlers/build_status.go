/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
)

// SetBuildStatus : sets the status of a build
func SetBuildStatus(msg *nats.Msg) {
	var err error
	var data []byte
	var e *models.Environment
	var b models.Build
	var cb *models.Build
	var bs struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	defer response(msg.Reply, &data, &err)

	err = json.Unmarshal(msg.Data, &bs)
	if err != nil {
		return
	}

	if bs.ID == "" && bs.Name == "" {
		data = []byte(`{"error": "not found"}`)
		return
	}

	if bs.ID == "" && bs.Name != "" {
		e, err = models.GetEnvironment(map[string]interface{}{"name": bs.Name})
		if err != nil {
			return
		}

		cb, err = models.GetLatestBuild(e.ID)
		if err != nil {
			return
		}

		bs.ID = cb.UUID
	}

	err = b.SetStatus(bs.ID, bs.Status)
	if err != nil {
		return
	}

	data = []byte(`{"status": "ok"}`)
}
