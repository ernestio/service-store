/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"errors"
	"time"

	"github.com/r3labs/graph"
)

// BuildFields ...
var BuildFields = append([]string{"uuid"}, structFields(Build{})...)

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
	query(q, BuildFields, []string{}).Find(&builds)
	return builds
}

// GetBuild ...
func GetBuild(q map[string]interface{}) (*Build, error) {
	var build Build
	err := query(q, BuildFields, []string{}).First(&build).Error
	return &build, err
}

// Create ...
func (b *Build) Create() error {
	return DB.Create(b).Error
}

// Update ...
func (b *Build) Update() error {
	var stored Build

	err := DB.Where("uuid = ?", b.UUID).First(&stored).Error
	if err != nil {
		return err
	}

	stored.Status = b.Status
	stored.Definition = b.Definition
	stored.Mapping = b.Mapping

	return DB.Save(&stored).Error
}

// Delete ...
func (b *Build) Delete() error {
	return DB.Delete(b).Error
}

// SetComponent : creates or updates a component
func (b *Build) SetComponent(c *graph.GenericComponent) error {
	var g *graph.Graph

	err := g.Load(b.Mapping)
	if err != nil {
		return err
	}

	if g.HasComponent(c.GetID()) {
		g.UpdateComponent(c)
	} else {
		err = g.AddComponent(c)
		if err != nil {
			return err
		}
	}

	b.Mapping.LoadGraph(g)

	return nil
}

// DeleteComponent : updates a component
func (b *Build) DeleteComponent(c *graph.GenericComponent) error {
	var g *graph.Graph

	err := g.Load(b.Mapping)
	if err != nil {
		return err
	}

	g.DeleteComponent(c)

	b.Mapping.LoadGraph(g)

	return nil
}

// SetChange : updates a change
func (b *Build) SetChange(c *graph.GenericComponent) error {
	var g *graph.Graph

	err := g.Load(b.Mapping)
	if err != nil {
		return err
	}

	for i := 0; i < len(g.Changes); i++ {
		if g.Changes[i].GetID() == c.GetID() {
			g.Changes[i] = c
			b.Mapping.LoadGraph(g)
			return nil
		}
	}

	return errors.New("change not found")
}

// DeleteChange : deletes a change
func (b *Build) DeleteChange(c *graph.GenericComponent) error {
	var g *graph.Graph

	err := g.Load(b.Mapping)
	if err != nil {
		return err
	}

	for i := len(g.Changes) - 1; i >= 0; i-- {
		if g.Changes[i].GetID() == c.GetID() {
			g.Changes = append(g.Changes[:i], g.Changes[i+1:]...)
		}
	}

	b.Mapping.LoadGraph(g)

	return nil
}
