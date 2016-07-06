/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"time"
	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// Entity : the database mapped entity
type Entity struct {
	ID           uint      `json:"-" gorm:"primary_key"`
	Uuid         string    `json:"id"`
	GroupID      uint      `json:"group_id"`
	DatacenterID uint      `json:"datacenter_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Version      time.Time `json:"version"`
	Status       string    `json:"status"`
	Options      string    `json:"options"`
	Definition   string    `json:"definition"`
	Mapping      string    `json:"mapping" gorm:"type:text;"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be services
func (Entity) TableName() string {
	return "services"
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (e *Entity) Find() []interface{} {
	entities := []Entity{}
	if e.Name != "" && e.GroupID != 0 {
		db.Where("name = ?", e.Name).Where("group_id = ?", e.GroupID).Order("version desc").Find(&entities)
	} else {
		if e.Name != "" {
			db.Where("name = ?", e.Name).Order("version desc").Find(&entities)
		} else if e.GroupID != 0 {
			db.Where("group_id = ?", e.GroupID).Order("version desc").Find(&entities)
		}
	}

	var list []interface{}
	prev := ""
	for _, s := range entities {
		if prev != s.Name {
			list = append(list, s)
		}
		prev = s.Name
	}

	return list
}

// MapInput : maps the input []byte on the current entity
func (e *Entity) MapInput(body []byte) {
	json.Unmarshal(body, &e)
}

// HasID : determines if the current entity has an id or not
func (e *Entity) HasID() bool {
	if e.ID == 0 {
		return false
	}
	return true
}

// LoadFromInput : Will load from a []byte input the database stored entity
func (e *Entity) LoadFromInput(msg []byte) bool {
	e.MapInput(msg)
	var stored Entity
	if e.Uuid != "" {
		db.Where("uuid = ?", e.Uuid).First(&stored)
	} else if e.Name != "" {
		db.Where("name = ?", e.Name).First(&stored)
	}
	if &stored == nil {
		return false
	}
	if ok := stored.HasID(); !ok {
		return false
	}
	e.Name = stored.Name
	e.Uuid = stored.Uuid
	e.GroupID = stored.GroupID
	e.DatacenterID = stored.DatacenterID
	e.Type = stored.Type
	e.Version = stored.Version
	e.Status = stored.Status
	e.Options = stored.Options
	e.Definition = stored.Definition
	e.Mapping = stored.Mapping
	e.ID = stored.ID

	return true
}

// LoadFromInputOrFail : Will try to load from the input an existing entity,
// or will call the handler to Fail the nats message
func (e *Entity) LoadFromInputOrFail(msg *nats.Msg, h *natsdb.Handler) bool {
	stored := &Entity{}
	ok := stored.LoadFromInput(msg.Data)
	if !ok {
		h.Fail(msg)
	}
	*e = *stored

	return ok
}

// Update : It will update the current entity with the input []byte
func (e *Entity) Update(body []byte) error {
	e.MapInput(body)
	stored := Entity{}
	db.Where("uuid = ?", e.Uuid).First(&stored)
	stored.Name = e.Name
	stored.Uuid = e.Uuid
	stored.GroupID = e.GroupID
	stored.DatacenterID = e.DatacenterID
	stored.Type = e.Type
	stored.Version = e.Version
	stored.Status = e.Status
	stored.Options = e.Options
	stored.Definition = e.Definition
	stored.Mapping = e.Mapping
	stored.ID = e.ID

	db.Save(&stored)
	e = &stored

	return nil
}

// Delete : Will delete from database the current Entity
func (e *Entity) Delete() error {
	db.Unscoped().Delete(&e)

	return nil
}

// Save : Persists current entity on database
func (e *Entity) Save() error {
	db.Save(&e)

	return nil
}
