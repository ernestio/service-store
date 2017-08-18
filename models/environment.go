/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import "time"

// EnvironmentFields ...
var EnvironmentFields = structFields(Environment{})

// EnvironmentQueryFields ...
var EnvironmentQueryFields = []string{"ids", "names"}

// Environment : the database mapped entity
type Environment struct {
	ID           uint       `json:"id" gorm:"primary_key"`
	GroupID      uint       `json:"group_id"`
	DatacenterID uint       `json:"datacenter_id"`
	Name         string     `json:"name" gorm:"type:varchar(100);unique_index"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Options      Map        `json:"option" gorm:"type: jsonb not null default '{}'::jsonb"`
	Credentials  Map        `json:"credentials" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be services
func (e *Environment) TableName() string {
	return "environments"
}

// FindEnvironments : finds a service
func FindEnvironments(q map[string]interface{}) []Environment {
	var services []Environment
	query(q, EnvironmentFields, EnvironmentQueryFields).Find(&services)
	return services
}

// GetEnvironment ....
func GetEnvironment(q map[string]interface{}) (*Environment, error) {
	var service Environment
	err := query(q, EnvironmentFields, EnvironmentQueryFields).First(&service).Error
	return &service, err
}

// Create ...
func (e *Environment) Create() error {
	return DB.Create(e).Error
}

// Update ...
func (e *Environment) Update() error {
	var stored Environment

	err := DB.Where("id = ?", e.ID).First(&stored).Error
	if err != nil {
		return err
	}

	stored.Options = e.Options

	return DB.Save(&stored).Error
}

// Delete ...
func (e *Environment) Delete() error {
	err := DB.Unscoped().Where("environment_id = ?", e.ID).Delete(Build{}).Error
	if err != nil {
		return err
	}

	return DB.Unscoped().Delete(e).Error
}
