/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import "time"

// EnvironmentFields ...
var EnvironmentFields = structFields(Environment{})

// EnvironmentQueryFields ...
var EnvironmentQueryFields = []string{"ids->id", "names->name"}

// Environment : the database mapped entity
type Environment struct {
	ID          uint       `json:"id" gorm:"primary_key"`
	ProjectID   uint       `json:"project_id"`
	Name        string     `json:"name" gorm:"type:varchar(100);unique_index"`
	Type        string     `json:"type"`
	Status      string     `json:"status"`
	Options     Map        `json:"option" gorm:"type: jsonb not null default '{}'::jsonb"`
	Credentials Map        `json:"credentials" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be environments
func (e *Environment) TableName() string {
	return "environments"
}

// FindEnvironments : finds a environment
func FindEnvironments(q map[string]interface{}) ([]Environment, error) {
	var environments []Environment
	err := query(q, EnvironmentFields, EnvironmentQueryFields).Find(&environments).Error
	return environments, err
}

// GetEnvironment ....
func GetEnvironment(q map[string]interface{}) (*Environment, error) {
	var environment Environment
	err := query(q, EnvironmentFields, EnvironmentQueryFields).First(&environment).Error
	return &environment, err
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

	if e.Options != nil {
		stored.Options = e.Options
	}
	if e.Credentials != nil {
		stored.Credentials = e.Credentials
	}

	return DB.Save(&stored).Error
}

// Delete ...
func (e *Environment) Delete() error {
	var err error

	if e.ID == 0 {
		err = DB.Where("name = ?", e.Name).First(e).Error
		if err != nil {
			return err
		}
	}

	err = DB.Unscoped().Where("environment_id = ?", e.ID).Delete(Build{}).Error
	if err != nil {
		return err
	}

	return DB.Unscoped().Delete(e).Error
}
