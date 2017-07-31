/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// ServiceView : renders an old singular service by joining data on builds and services tables
type ServiceView struct {
	ID           uint       `json:"-" gorm:"primary_key"`
	IDs          []string   `json:"ids" gorm:"-"`
	Names        []string   `json:"names" gorm:"-"`
	UUID         string     `json:"id"`
	GroupID      uint       `json:"group_id"`
	UserID       uint       `json:"user_id"`
	DatacenterID uint       `json:"datacenter_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Version      time.Time  `json:"version"`
	Status       string     `json:"status"`
	Options      models.Map `json:"options"`
	Definition   string     `json:"definition"`
	Mapping      models.Map `json:"mapping" gorm:"type:text;"`
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (s *ServiceView) Find() []interface{} {
	var results []interface{}
	var services []ServiceView

	q := db.Table("services").Select("builds.id as id, builds.uuid, builds.user_id, builds.status, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options").Joins("INNER JOIN builds ON (builds.service_id = services.id)")

	if s.Name != "" && s.GroupID != 0 {
		if s.UUID != "" {
			q = q.Where("services.name = ?", s.Name).Where("services.group_id = ?", s.GroupID).Where("builds.uuid = ?", s.UUID)
		} else {
			q = q.Where("services.name = ?", s.Name).Where("services.group_id = ?", s.GroupID)
		}
	} else {
		if s.UUID != "" {
			q = q.Where("builds.id = ?", s.UUID)
		} else if s.Name != "" {
			q = q.Where("services.name = ?", s.Name)
		} else if s.GroupID != 0 {
			q = q.Where("services.group_id = ?", s.GroupID)
		} else if s.DatacenterID != 0 {
			q = q.Where("services.datacenter_id = ?", s.DatacenterID)
		}
	}

	q.Order("version desc").Find(&services)

	results = make([]interface{}, len(services))

	for i := 0; i < len(services); i++ {
		results[i] = &services[i]
	}

	return results
}

// MapInput : maps the input []byte on the current entity
func (s *ServiceView) MapInput(body []byte) {
	if err := json.Unmarshal(body, &s); err != nil {
		log.Println(err)
	}
}

// HasID : determines if the current entity has an id or not
func (s *ServiceView) HasID() bool {
	if s.ID == 0 {
		return false
	}
	return true
}

// LoadFromInput : Will load from a []byte input the database stored entity
func (s *ServiceView) LoadFromInput(msg []byte) bool {
	s.MapInput(msg)
	var stored ServiceView

	q := db.Table("services").Select("builds.id as id, builds.uuid, builds.user_id, builds.status, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options").Joins("INNER JOIN builds ON (builds.service_id = services.id)")

	if s.UUID != "" {
		q = q.Where("builds.uuid = ?", s.UUID)
	} else if s.Name != "" {
		q = q.Where("services.name = ?", s.Name)
	}

	q.First(&stored)

	if &stored == nil {
		return false
	}

	if stored.HasID() != true {
		return false
	}

	s.Name = stored.Name
	s.UUID = stored.UUID
	s.GroupID = stored.GroupID
	s.UserID = stored.UserID
	s.DatacenterID = stored.DatacenterID
	s.Type = stored.Type
	s.Version = stored.Version
	s.Status = stored.Status
	s.Options = stored.Options
	s.Definition = stored.Definition
	s.Mapping = stored.Mapping
	s.ID = stored.ID

	return true
}

// LoadFromInputOrFail : Will try to load from the input an existing entity,
// or will call the handler to Fail the nats message
func (s *ServiceView) LoadFromInputOrFail(msg *nats.Msg, h *natsdb.Handler) bool {
	stored := &ServiceView{}
	ok := stored.LoadFromInput(msg.Data)
	if !ok {
		h.Fail(msg)
	}
	*s = *stored

	return ok
}

// Update : It will update the current entity with the input []byte
func (s *ServiceView) Update(body []byte) error {
	s.MapInput(body)

	if s.Name == "" {
		return nil
	}

	service := models.Service{
		Options: s.Options,
		Status:  s.Status,
	}

	err := service.Update()
	if err != nil {
		return err
	}

	build := models.Build{
		Status:     s.Status,
		Definition: s.Definition,
		Mapping:    s.Mapping,
	}

	db.Where("name = ?", s.Name).First(&service)

	db.Save(&service)

	db.Where("uuid = ?", s.UUID).First(&build)

	build.ServiceID = service.ID
	build.UserID = s.UserID
	build.Type = s.Type

	db.Save(&build)

	s.ID = service.ID

	return nil
}

// Delete : Will delete from database the current ServiceView
func (s *ServiceView) Delete() error {
	var service models.Service

	db.Where("name = ?", s.Name).First(&service)
	db.Unscoped().Where("id = ?", s.ID).Delete(&service)
	db.Unscoped().Where("service_id = ?", s.ID).Delete(models.Build{})

	return nil
}

// Save : Persists current entity on database
func (s *ServiceView) Save() error {
	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	service := models.Service{
		Name:         s.Name,
		GroupID:      s.GroupID,
		DatacenterID: s.DatacenterID,
		Options:      s.Options,
		Status:       s.Status,
	}

	err := tx.Save(&service).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	build := models.Build{
		UUID:       s.UUID,
		ServiceID:  service.ID,
		UserID:     s.UserID,
		Type:       s.Type,
		Status:     s.Status,
		Definition: s.Definition,
		Mapping:    s.Mapping,
	}

	err = tx.Save(&build).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

/*

func (e *ServiceView) requestDefinition() string {
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
func (e *ServiceView) setComponent(xc map[string]interface{}) error {
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
func (e *ServiceView) getComponent(id string) (*map[string]interface{}, error) {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return nil, err
	}

	return m.GetComponent(id)
}

// DeleteComponent : deletes a component from the mapping based on id
func (e *ServiceView) deleteComponent(id string) error {
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
func (e *ServiceView) setChange(xc map[string]interface{}) error {
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
func (e *ServiceView) getChange(id string) (*map[string]interface{}, error) {
	var m Mapping

	err := m.Load([]byte(e.Mapping))
	if err != nil {
		return nil, err
	}

	return m.GetChange(id)
}

// DeleteChange : deletes a change from the mapping based on id
func (e *ServiceView) deleteChange(id string) error {
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


*/
