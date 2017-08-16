/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
	"github.com/r3labs/natsdb"
)

// ServiceView : renders an old singular service by joining data on builds and services tables
type ServiceView struct {
	ID           uint       `json:"-" gorm:"primary_key"`
	IDs          []string   `json:"ids,omitempty" gorm:"-"`
	Names        []string   `json:"names,omitempty" gorm:"-"`
	UUID         string     `json:"id"`
	GroupID      uint       `json:"group_id"`
	UserID       uint       `json:"user_id"`
	DatacenterID uint       `json:"datacenter_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Version      time.Time  `json:"version"`
	Status       string     `json:"status"`
	Options      models.Map `json:"options"`
	Credentials  models.Map `json:"credentials"`
	Definition   string     `json:"definition,omitempty"`
	Mapping      models.Map `json:"mapping,omitempty" gorm:"type:text;"`
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (s *ServiceView) Find() []interface{} {
	var results []interface{}
	var services []ServiceView

	q := db.Table("services").Select("builds.id as id, builds.uuid, builds.user_id, builds.status, builds.definition, builds.created_at as version, services.name, services.group_id, services.datacenter_id, services.options, services.credentials, services.type").Joins("INNER JOIN builds ON (builds.service_id = services.id)")

	if len(s.IDs) > 0 {
		q = q.Where("builds.uuid in (?)", s.IDs)
	} else if len(s.Names) > 0 {
		q = q.Where("services.name in (?)", s.Names)
	} else if s.Name != "" && s.GroupID != 0 {
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
	return s.ID != 0
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

	err := q.First(&stored).Error
	if err != nil {
		return false
	}

	if !stored.HasID() {
		return false
	}

	*s = stored

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
		return errors.New("service name was not specified")
	}

	service := models.Service{
		Options:     s.Options,
		Credentials: s.Credentials,
	}

	db.Where("name = ?", s.Name).First(&service)
	return db.Save(&service).Error
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
	var err error

	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	defer func() {
		switch err {
		case nil:
			err = tx.Commit().Error
		default:
			log.Println(err)
			err = tx.Rollback().Error
		}
	}()

	service := models.Service{
		Name:         s.Name,
		GroupID:      s.GroupID,
		DatacenterID: s.DatacenterID,
		Type:         s.Type,
		Options:      s.Options,
		Credentials:  s.Credentials,
		Status:       "initializing",
	}

	err = tx.Where("name = ?", s.Name).FirstOrCreate(&service).Error
	if err != nil {
		return err
	}

	switch service.Status {
	case "initializing", "done", "errored":
		err = tx.Exec("UPDATE services SET status = ? WHERE id = ?", "in_progress", service.ID).Error
	case "in_progress":
		err = errors.New(`{"error": "could not create service build: service in progress"}`)
	default:
		err = errors.New(`{"error": "could not create service build: unknown service state"}`)
	}

	if err != nil {
		return err
	}

	build := models.Build{
		UUID:       s.UUID,
		ServiceID:  service.ID,
		UserID:     s.UserID,
		Type:       s.Type,
		Status:     "in_progress",
		Definition: s.Definition,
		Mapping:    s.Mapping,
	}

	err = tx.Save(&build).Error
	if err != nil {
		return err
	}

	s.Version = build.CreatedAt
	s.Status = build.Status

	return nil
}
