/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"encoding/json"
	"log"
	"time"
)

// Service : the database mapped entity
type Service struct {
	ID           uint   `json:"id" gorm:"primary_key"`
	GroupID      uint   `json:"group_id"`
	DatacenterID uint   `json:"datacenter_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Sync         bool   `json:"sync"`
	SyncType     string `json:"sync_type"`
	SyncInterval int    `json:"sync_interval"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// Find : based on the defined fields for the current entity
// will perform a search on the database
func (s *Service) Find() []Service {
	var services []Service

	if s.Name != "" && s.GroupID != 0 {
		if s.ID != 0 {

			DB.Where("name = ?", s.Name).Where("group_id = ?", s.GroupID).Where("id = ?", s.ID).Find(&services)
		} else {
			DB.Where("name = ?", s.Name).Where("group_id = ?", s.GroupID).Find(&services)
		}
	} else {
		if s.Name != "" && s.ID != 0 {
			DB.Where("name = ?", s.Name).Where("id = ?", s.ID).Find(&services)
		} else if s.Name != "" {
			DB.Where("name = ?", s.Name).Find(&services)
		} else if s.GroupID != 0 {
			DB.Where("group_id = ?", s.GroupID).Find(&services)
		} else if s.DatacenterID != 0 {
			DB.Where("datacenter_id = ?", s.DatacenterID).Find(&services)
		}
	}

	return services
}

// MapInput : maps the input []byte on the current entity
func (s *Service) MapInput(body []byte) {
	if err := json.Unmarshal(body, &e); err != nil {
		log.Println(err)
	}
}

// HasID : determines if the current entity has an id or not
func (s *Service) HasID() bool {
	if s.ID == 0 {
		return false
	}
	return true
}

// LoadFromInput : Will load from a []byte input the database stored entity
func (s *Service) Load(data []byte) bool {
	var stored Service

	s.MapInput(msg)
	if s.ID != 0 {
		DB.Where("id = ?", s.ID).First(&stored)
	} else if s.Name != "" {
		DB.Where("name = ?", s.Name).First(&stored)
	}
	if &stored == nil {
		return false
	}
	if ok := stored.HasID(); !ok {
		return false
	}
	s.ID = stored.ID
	s.Name = stored.Name
	s.GroupID = stored.GroupID
	s.DatacenterID = stored.DatacenterID
	s.Type = stored.Type
	s.Status = stored.Status
	s.Sync = stored.Sync
	s.SyncType = stored.SyncType
	s.SyncInterval = stored.SyncInterval

	return true
}

// Update : It will update the current entity with the input []byte
func (s *Service) Update(body []byte) error {
	s.MapInput(body)
	stored := Service{}
	DB.Where("id = ?", s.ID).First(&stored)
	stored.ID = s.ID
	stored.Name = s.Name
	stored.GroupID = s.GroupID
	stored.DatacenterID = s.DatacenterID
	stored.Type = s.Type
	stored.Status = s.Status
	stored.Sync = s.Sync
	stored.SyncType = s.SyncType
	stored.SyncInterval = s.SyncInterval

	DB.Save(&stored)
	e = &stored

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

	err := tx.Save(e).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}
