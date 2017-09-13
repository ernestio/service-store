/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import "encoding/json"

// Error : default error message
type Error struct {
	Error string `json:"error"`
}

func errResponse(subject string, err error) {
	if err != nil {
		data, _ := json.Marshal(Error{Error: err.Error()})
		NC.Publish(subject, data)
	}
}

func response(subject string, data []byte, err error) {
	if err != nil {
		data, _ = json.Marshal(Error{Error: err.Error()})
	}

	NC.Publish(subject, data)
}
