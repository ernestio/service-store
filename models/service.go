/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"time"
)

// ServiceFields ...
var ServiceFields = structFields(Service{})

// ServiceQueryFields ...
var ServiceQueryFields = []string{"ids", "names"}

// Service : the database mapped entity
type Service struct {
	ID           uint       `json:"id" gorm:"primary_key"`
	GroupID      uint       `json:"group_id"`
	DatacenterID uint       `json:"datacenter_id"`
	Name         string     `json:"name" gorm:"type:varchar(100);unique_index"`
	Status       string     `json:"status"`
	Options      Map        `json:"option" gorm:"type: jsonb not null default '{}'::jsonb"`
	Credentials  Map        `json:"credentials" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be services
func (s *Service) TableName() string {
	return "services"
}

// FindServices : finds a service
func FindServices(q map[string]interface{}) []Service {
	var services []Service
	query(q, ServiceFields, ServiceQueryFields).Find(&services)
	return services
}

// GetService ....
func GetService(q map[string]interface{}) (*Service, error) {
	var service *Service
	err := query(q, ServiceFields, ServiceQueryFields).First(service).Error
	return service, err
}

// Create ...
func (s *Service) Create() error {
	return DB.Create(s).Error
}

// Update ...
func (s *Service) Update() error {
	var stored Service

	err := DB.Where("id = ?", s.ID).First(&stored).Error
	if err != nil {
		return err
	}

	stored.Options = s.Options
	stored.Status = s.Status

	return DB.Save(&stored).Error
}

// Delete ...
func (s *Service) Delete() error {
	err := DB.Unscoped().Where("service_id = ?", s.ID).Delete(Build{}).Error
	if err != nil {
		return err
	}

	return DB.Unscoped().Delete(s).Error
}
