/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"os"
	"time"

	aes "github.com/ernestio/crypto/aes"
)

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
	Options     Map        `json:"options" gorm:"type: jsonb not null default '{}'::jsonb"`
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
	err := query(q, EnvironmentFields, EnvironmentQueryFields).Order("updated_at desc").Find(&environments).Error
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
	ec, err := encryptCredentials(e.Credentials)
	if err != nil {
		return err
	}

	e.Credentials = ec

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
		ec, err := encryptCredentials(e.Credentials)
		if err != nil {
			return err
		}

		stored.Credentials = ec
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

// GetState ...
func (e *Environment) GetState() string {
	return e.Status
}

// SetState ...
func (e *Environment) SetState(state string) {
	e.Status = state
}

func crypt(s string) (string, error) {
	crypto := aes.New()
	key := os.Getenv("ERNEST_CRYPTO_KEY")
	if s != "" {
		encrypted, err := crypto.Encrypt(s, key)
		if err != nil {
			return "", err
		}
		s = encrypted
	}

	return s, nil
}

func encryptCredentials(c Map) (Map, error) {
	for k, v := range c {
		if k == "region" || k == "external_network" || k == "username" || k == "vcloud_url" {
			continue
		}

		xc, ok := v.(string)
		if !ok {
			continue
		}

		x, err := crypt(xc)
		if err != nil {
			return c, err
		}

		c[k] = x
	}

	return c, nil
}
