/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// Entity : the database mapped entity
type Entity struct {
	ID             uint      `json:"-" gorm:"primary_key"`
	UUID           string    `json:"id"`
	GroupID        uint      `json:"group_id"`
	UserID         uint      `json:"user_id"`
	DatacenterID   uint      `json:"datacenter_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Version        time.Time `json:"version"`
	Status         string    `json:"status"`
	LastKnownError string    `json:"last_known_error"`
	Options        string    `json:"options"`
	Definition     string    `json:"definition"`
	Endpoint       string    `json:"endpoint" gorm:"-"`
	Mapping        string    `json:"mapping" gorm:"type:text;"`
	Sync           bool      `json:"sync"`
	SyncType       string    `json:"sync_type"`
	SyncInterval   int       `json:"sync_interval"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be services
func (Entity) TableName() string {
	return "services"
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (e *Entity) Find() []interface{} {
	entities := []Entity{}
	fields := "uuid, group_id, user_id, datacenter_id, name, type, version, status, options, definition, mapping, last_known_error"
	if e.Name != "" && e.GroupID != 0 {
		if e.UUID != "" {
			db.Select(fields).Where("name = ?", e.Name).Where("group_id = ?", e.GroupID).Where("uuid = ?", e.UUID).Order("version desc").Find(&entities)
		} else {
			db.Select(fields).Where("name = ?", e.Name).Where("group_id = ?", e.GroupID).Order("version desc").Find(&entities)
		}
	} else {
		if e.Name != "" && e.UUID != "" {
			db.Select(fields).Where("name = ?", e.Name).Where("uuid = ?", e.UUID).Order("version desc").Find(&entities)
		} else if e.Name != "" {
			db.Select(fields).Where("name = ?", e.Name).Order("version desc").Find(&entities)
		} else if e.GroupID != 0 {
			db.Select(fields).Where("group_id = ?", e.GroupID).Order("version desc").Find(&entities)
		} else if e.DatacenterID != 0 {
			db.Select(fields).Where("datacenter_id = ?", e.DatacenterID).Order("version desc").Find(&entities)
		}
	}

	list := make([]interface{}, len(entities))
	for i, s := range entities {
		s.Endpoint = s.getEndpoint()
		s.Mapping = ""
		list[i] = s
	}

	return list
}

func (e *Entity) getEndpoint() string {
	var s struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.Unmarshal([]byte(e.Mapping), &s); err != nil {
		log.Println(err)
	}

	return s.Endpoint

}

// MapInput : maps the input []byte on the current entity
func (e *Entity) MapInput(body []byte) {
	if err := json.Unmarshal(body, &e); err != nil {
		log.Println(err)
	}
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
	if e.UUID != "" {
		db.Where("uuid = ?", e.UUID).First(&stored)
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
	e.UUID = stored.UUID
	e.GroupID = stored.GroupID
	e.UserID = stored.UserID
	e.DatacenterID = stored.DatacenterID
	e.Type = stored.Type
	e.Version = stored.Version
	e.Status = stored.Status
	e.Sync = stored.Sync
	e.SyncType = stored.SyncType
	e.SyncInterval = stored.SyncInterval
	e.LastKnownError = stored.LastKnownError
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
	db.Where("uuid = ?", e.UUID).First(&stored)
	stored.Name = e.Name
	stored.UUID = e.UUID
	stored.GroupID = e.GroupID
	stored.UserID = e.UserID
	stored.DatacenterID = e.DatacenterID
	stored.Type = e.Type
	stored.Version = e.Version
	if e.Status == "done" && e.Status != stored.Status {
		stored.Definition = e.requestDefinition()
	} else {
		stored.Definition = e.Definition
	}
	stored.Status = e.Status
	stored.LastKnownError = e.LastKnownError
	stored.Sync = e.Sync
	stored.SyncType = e.SyncType
	stored.SyncInterval = e.SyncInterval
	stored.Options = e.Options
	stored.Mapping = e.Mapping
	stored.ID = e.ID

	db.Save(&stored)
	e = &stored

	return nil
}

// Delete : Will delete from database the current Entity
func (e *Entity) Delete() error {
	db.Unscoped().Where("name = ?", e.Name).Delete(Entity{})

	return nil
}

// Save : Persists current entity on database
func (e *Entity) Save() error {
	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	err := tx.Save(e).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (e *Entity) requestDefinition() string {
	body, err := json.Marshal(e)
	if err != nil {
		log.Panic(err)
	}
	res, err := n.Request("definition.map.service", body, time.Second)
	if err != nil {
		log.Panic(err)
	}
	return string(res.Data)
}

// SetComponent : sets a component on a services mapping
func (e *Entity) setComponent(xc map[string]interface{}) error {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return err
	}

	err = m.SetComponent(xc)
	if err != nil {
		return err
	}

	data, err := m.ToJSON()
	if err != nil {
		return err
	}

	e.Mapping = string(data)

	return nil
}

// GetComponent : returns a component from a services mapping based on it's id
func (e *Entity) getComponent(id string) (*map[string]interface{}, error) {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return nil, err
	}

	return m.GetComponent(id)
}

// DeleteComponent : deletes a component from the mapping based on id
func (e *Entity) deleteComponent(id string) error {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return err
	}

	err = m.DeleteComponent(id)
	if err != nil {
		return err
	}

	data, err := m.ToJSON()
	if err != nil {
		return err
	}

	e.Mapping = string(data)

	return nil
}

// SetChange : sets a change on a services mapping
func (e *Entity) setChange(xc map[string]interface{}) error {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return err
	}

	err = m.SetChange(xc)
	if err != nil {
		return err
	}

	data, err := m.ToJSON()
	if err != nil {
		return err
	}

	e.Mapping = string(data)

	return nil
}

// GetChange : returns a change from a services mapping based on it's id
func (e *Entity) getChange(id string) (*map[string]interface{}, error) {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return nil, err
	}

	return m.GetChange(id)
}

// DeleteChange : deletes a change from the mapping based on id
func (e *Entity) deleteChange(id string) error {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return err
	}

	err = m.DeleteChange(id)
	if err != nil {
		return err
	}

	data, err := m.ToJSON()
	if err != nil {
		return err
	}

	e.Mapping = string(data)

	return nil
}
