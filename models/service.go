/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"encoding/json"
	"errors"
	"log"
	"time"
)

// Service : the database mapped entity
type Service struct {
	ID           uint       `json:"id" gorm:"primary_key"`
	GroupID      uint       `json:"group_id"`
	DatacenterID uint       `json:"datacenter_id"`
	Name         string     `json:"name" gorm:"type:varchar(100);unique_index"`
	Status       string     `json:"status"`
	Options      Map        `json:"option" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be services
func (s *Service) TableName() string {
	return "services"
}

// HasID : determines if the current entity has an id or not
func (s *Service) HasID() bool {
	return s.ID != 0
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (s *Service) Find() []Service {
	var services []Service

	fields := "id, group_id, datacenter_id, name, status, options"

	if s.Name != "" && s.GroupID != 0 {
		if s.ID != 0 {
			DB.Select(fields).Where("name = ?", s.Name).Where("group_id = ?", s.GroupID).Where("id = ?", s.ID).Find(&services)
		} else {
			DB.Select(fields).Where("name = ?", s.Name).Where("group_id = ?", s.GroupID).Find(&services)
		}
	} else {
		if s.Name != "" && s.ID != 0 {
			DB.Select(fields).Where("name = ?", s.Name).Where("id = ?", s.ID).Find(&services)
		} else if s.Name != "" {
			DB.Select(fields).Where("name = ?", s.Name).Find(&services)
		} else if s.GroupID != 0 {
			DB.Select(fields).Where("group_id = ?", s.GroupID).Find(&services)
		} else if s.DatacenterID != 0 {
			DB.Select(fields).Where("datacenter_id = ?", s.DatacenterID).Find(&services)
		}
	}

	return services
}

// MapInput : maps the input []byte on the current entity
func (s *Service) MapInput(body []byte) error {
	return json.Unmarshal(body, s)
}

// Load : Will load from a []byte input the database stored entity
func (s *Service) Load(data []byte) error {
	var stored Service

	err := s.MapInput(data)
	if err != nil {
		return err
	}

	if s.ID != 0 {
		DB.Where("id = ?", s.ID).First(&stored)
	} else if s.Name != "" {
		DB.Where("name = ?", s.Name).First(&stored)
	}

	if &stored == nil {
		return errors.New("could not find service")
	}

	if stored.HasID() != true {
		return errors.New("stored component has no id")
	}

	s.ID = stored.ID
	s.Name = stored.Name
	s.GroupID = stored.GroupID
	s.DatacenterID = stored.DatacenterID
	s.Status = stored.Status
	s.Options = stored.Options

	return nil
}

// Update : It will update the current entity with the input []byte
func (s *Service) Update(body []byte) error {
	err := s.MapInput(body)
	if err != nil {
		return err
	}

	stored := Service{}
	DB.Where("id = ?", s.ID).First(&stored)
	stored.ID = s.ID
	stored.Name = s.Name
	stored.GroupID = s.GroupID
	stored.DatacenterID = s.DatacenterID
	stored.Status = s.Status
	stored.Options = s.Options

	DB.Save(&stored)
	s = &stored

	return nil
}

// Delete : Will delete from database the current Service
func (s *Service) Delete() error {
	DB.Unscoped().Where("name = ?", s.Name).Delete(Service{})

	return nil
}

// Save : Persists current entity on database
func (s *Service) Save() error {
	tx := DB.Begin()
	tx.Exec("set transaction isolation level serializable")

	err := tx.Save(s).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
