/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package handlers

import (
	"encoding/json"
	"log"
	"time"
)

// Error : default error message
type Error struct {
	Error string `json:"_error"`
}

// Message ...
type Message struct {
	ID         string                 `json:"id"`
	Definition string                 `json:"definition"`
	Mapping    map[string]interface{} `json:"mapping"`
	Validation map[string]interface{} `json:"validation"`
}

func response(reply string, data *[]byte, err *error) {
	var rdata []byte
	if data != nil {
		rdata = *data
	}

	if *err != nil {
		log.Println("[ ERROR ] " + (*err).Error())
		rdata, _ = json.Marshal(Error{Error: (*err).Error()})
	}

	if reply != "" {
		NC.Publish(reply, rdata)
	}
}

func pub(subject string, data []byte) {
	if err := NC.Publish(subject, data); err != nil {
		log.Println("[ERROR] : " + err.Error())
	}
}

// DetatchPolicies : will detach all policies from an environment
func DetatchPolicies(env string) {
	var p []map[string]interface{}

	resp, err := NC.Request("policy.find", []byte(`{"environments": ["`+env+`"]}`), time.Second*5)
	if err != nil {
		log.Println("[ERROR] : " + err.Error())
		return
	}

	err = json.Unmarshal(resp.Data, &p)
	if err != nil {
		log.Println("[ERROR] : " + err.Error())
		return
	}

	if len(p) < 1 {
		return
	}

	for i := 0; i < len(p); i++ {
		if p[i]["environments"] == nil {
			continue
		}

		envs := p[i]["environments"].([]string)

		for x := len(envs) - 1; x >= 0; x-- {
			if envs[i] == env {
				envs = append(envs[:x], envs[x+1:]...)
			}
		}

		p[i]["environments"] = envs

		data, err := json.Marshal(p[i])
		if err != nil {
			log.Println("[ERROR] : " + err.Error())
			return
		}

		_, err = NC.Request("policy.set", data, time.Second*5)
		if err != nil {
			log.Println("[ERROR] : " + err.Error())
		}
	}
}
