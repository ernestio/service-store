/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"time"
)

// BuildFields ...
var BuildFields = structFields(Build{})

// Build : stores build data
type Build struct {
	ID         uint       `json:"-" gorm:"primary_key"`
	UUID       string     `json:"id"`
	ServiceID  uint       `json:"service_id" gorm:"ForeignKey:UserRefer"`
	UserID     uint       `json:"user_id"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	Definition string     `json:"definition" gorm:"type:text;"`
	Mapping    Map        `json:"mapping" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"-" sql:"index"`
}

// TableName : set Entity's table name to be builds
func (b *Build) TableName() string {
	return "builds"
}

// FindBuilds : finds a build
func FindBuilds(q map[string]interface{}) []Build {
	var builds []Build
	query(q, BuildFields).Find(&builds)
	return builds
}

// GetBuild ...
func GetBuild(q map[string]interface{}) Build {
	var build Build
	query(q, BuildFields).First(build)
	return build
}

// Create ...
func (b *Build) Create() error {
	return DB.Create(b).Error
}

// Update ...
func (b *Build) Update() error {
	var stored *Build

	err := DB.Where("uuid = ?", b.UUID).First(stored).Error
	if err != nil {
		return err
	}

	stored.Status = b.Status
	stored.Definition = b.Definition
	stored.Mapping = b.Mapping

	return DB.Save(stored).Error
}

// Delete ...
func (b *Build) Delete() error {
	return DB.Delete(b).Error
}
