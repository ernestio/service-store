/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// Component : represents a component from a service mapping
type Component map[string]interface{}

// MapInput : maps the input []byte on the current component
func (c *Component) MapInput(body []byte) {
	if err := json.Unmarshal(body, &c); err != nil {
		log.Println(err)
	}
}

// GetServiceID : gets the components service id
func (c *Component) GetServiceID() (string, error) {
	id, ok := (*c)["service"].(string)
	if !ok {
		return "", errors.New("could not get service id from component")
	}

	return id, nil
}

// GetComponentID : gets the components id
func (c *Component) GetComponentID() (string, error) {
	id, ok := (*c)["_component_id"].(string)
	if !ok {
		return "", errors.New("could not get component id from component")
	}

	return id, nil
}

// LoadFromInputOrFail : Will try to load from the input an existing component,
// or will call the handler to Fail the nats message
func (c *Component) LoadFromInputOrFail(msg *nats.Msg, h *natsdb.Handler) bool {
	c.MapInput(msg.Data)

	_, err := c.GetServiceID()
	if err != nil {
		log.Println(err)
		h.Fail(msg)
		return false
	}

	_, err = c.GetComponentID()
	if err != nil {
		log.Println(err)
		h.Fail(msg)
		return false
	}

	return true
}
