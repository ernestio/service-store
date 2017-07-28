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
	services := []*ServiceView{}

	db.LogMode(true)
	//q := db.Raw("SELECT builds.uuid, builds.user_id, builds.status, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options FROM services INNER JOIN builds ON (builds.service_id = services.id)")
	q := db.Table("services").Select("builds.uuid, builds.user_id, builds.status, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options").Joins("INNER JOIN builds ON (builds.service_id = services.id)")

	if s.Name != "" && s.GroupID != 0 {
		if s.UUID != "" {
			q.Where("service.name = ?", s.Name).Where("service.group_id = ?", s.GroupID).Where("builds.uuid = ?", s.UUID).Order("version desc").Find(&services)
		} else {
			q.Where("service.name = ?", s.Name).Where("service.group_id = ?", s.GroupID).Order("version desc").Find(&services)
		}
	} else {
		if s.Name != "" && s.UUID != "" {
			q.Where("service.name = ?", s.Name).Where("builds.uuid = ?", s.UUID).Order("version desc").Find(&services)
		} else if s.Name != "" {
			q.Where("service.name = ?", s.Name).Order("version desc").Find(&services)
		} else if s.GroupID != 0 {
			q.Where("service.group_id = ?", s.GroupID).Order("version desc").Find(&services)
		} else if s.DatacenterID != 0 {
			q.Where("service.datacenter_id = ?", s.DatacenterID).Order("version desc").Find(&services)
		}
	}
	db.LogMode(false)

	return nil
}

// MapInput : maps the input []byte on the current entity
func (e *ServiceView) MapInput(body []byte) {
	if err := json.Unmarshal(body, &e); err != nil {
		log.Println(err)
	}
}

// HasID : determines if the current entity has an id or not
func (e *ServiceView) HasID() bool {
	if e.ID == 0 {
		return false
	}
	return true
}

// LoadFromInput : Will load from a []byte input the database stored entity
func (s *ServiceView) LoadFromInput(msg []byte) bool {
	s.MapInput(msg)
	var stored ServiceView

	db.LogMode(true)
	q := db.Table("services").Select("builds.uuid, builds.user_id, builds.status, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options").Joins("INNER JOIN builds ON (builds.service_id = services.id)")

	if s.UUID != "" {
		q.Where("builds.uuid = ?", s.UUID).First(&stored)
	} else if s.Name != "" {
		q.Where("service.name = ?", s.Name).First(&stored)
	}
	db.LogMode(false)

	if &stored == nil {
		return false
	}
	if ok := stored.HasID(); !ok {
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
		log.Println("no service name specified!")
		return nil
	}

	service := models.Service{
		Options: s.Options,
		Status: s.Status,
	}

	err := service.Update()
	if err != nil {
		return err
	}

	build := models.Build{
		Status: s.Status,
		Definition: s.Definition,
		Mapping: s.Mapping,
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
	db.Unscoped().Where("name = ?", s.Name).Delete(ServiceView{})

	return nil
}

// Save : Persists current entity on database
func (s *ServiceView) Save() error {
	panic("saving!")

	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	service := models.Service{
		Name:
	}

	err := tx.Save(s).Error
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
