/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
)

// Mapping : holds the mapping for a service
type Mapping struct {
	ID         string                   `json:"id"`
	Action     string                   `json:"action"`
	Components []map[string]interface{} `json:"components"`
	Changes    []map[string]interface{} `json:"changes"`
	Edges      []map[string]interface{} `json:"edges"`
}

// Load : Load a mapping from json
func (m *Mapping) Load(data []byte) error {
	return json.Unmarshal(data, m)
}

// ToJSON : Marshal mapping to json
func (m *Mapping) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// ComponentIndex : returns the index of a component
func (m *Mapping) ComponentIndex(id string) *int {
	return getIndex(m.Components, id)
}

// ChangeIndex : returns the index of a change
func (m *Mapping) ChangeIndex(id string) *int {
	return getIndex(m.Changes, id)
}

// GetComponent : gets a component
func (m *Mapping) GetComponent(id string) (*map[string]interface{}, error) {
	index := m.ComponentIndex(id)
	if index == nil {
		return nil, errors.New("could not find component")
	}

	return &m.Components[*index], nil
}

// GetChange : gets a component
func (m *Mapping) GetChange(id string) (*map[string]interface{}, error) {
	index := m.ChangeIndex(id)
	if index == nil {
		return nil, errors.New("could not find change")
	}

	return &m.Changes[*index], nil
}

// SetComponent : sets a component on the mapping
func (m *Mapping) SetComponent(c map[string]interface{}) error {
	id := componentID(c)
	if id == nil {
		return errors.New("could not process component")
	}

	index := m.ComponentIndex(*id)
	if index != nil {
		m.Components[*index] = c
	} else {
		m.Components = append(m.Components, c)
	}

	return nil
}

// SetChange : sets a component on the mapping
func (m *Mapping) SetChange(c map[string]interface{}) error {
	id := componentID(c)
	if id == nil {
		return errors.New("could not process component")
	}

	index := m.ChangeIndex(*id)
	if index != nil {
		m.Changes[*index] = c
	} else {
		m.Changes = append(m.Changes, c)
	}

	return nil
}

// DeleteComponent : sets a component on the mapping
func (m *Mapping) DeleteComponent(id string) error {
	index := m.ComponentIndex(id)
	if index != nil {
		m.Components = append(m.Components[:*index], m.Components[*index+1:]...)
	}

	return nil
}

// DeleteChange : sets a component on the mapping
func (m *Mapping) DeleteChange(id string) error {
	index := m.ChangeIndex(id)
	if index != nil {
		m.Changes = append(m.Changes[:*index], m.Changes[*index+1:]...)
	}

	return nil
}

func getIndex(s []map[string]interface{}, id string) *int {
	for i := 0; i < len(s); i++ {
		cid := componentID(s[i])
		if cid == nil {
			continue
		}

		if *cid == id {
			return &i
		}
	}

	return nil
}

func componentID(c map[string]interface{}) *string {
	cid, ok := c["_component_id"].(string)
	if ok != true {
		return nil
	}

	return &cid
}
