/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"time"
)

var ServiceFields = structFields(Service{})

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
	query(q, ServiceFields).Find(&services)
	return services
}
