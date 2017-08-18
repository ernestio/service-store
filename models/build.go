/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"errors"
	"log"
	"time"

	graph "gopkg.in/r3labs/graph.v2"
)

// BuildFields ...
var BuildFields = append([]string{"uuid"}, structFields(Build{})...)

// GraphTransform : a function that can transform parts of a graph
type GraphTransform func(g *graph.Graph, c *graph.GenericComponent) error

// Build : stores build data
type Build struct {
	ID            uint       `json:"-" gorm:"primary_key"`
	UUID          string     `json:"id"`
	EnvironmentID uint       `json:"environment_id" gorm:"ForeignKey:UserRefer"`
	UserID        uint       `json:"user_id"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Definition    string     `json:"definition" gorm:"type:text;"`
	Mapping       Map        `json:"mapping" gorm:"type: jsonb not null default '{}'::jsonb"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"-" sql:"index"`
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

// GetLatestBuild : gets the latest build of a environment
func GetLatestBuild(environmentid uint) (*Build, error) {
	var build Build
	q := map[string]interface{}{"environment_id": environmentid}
	err := query(q, BuildFields, []string{}).Order("created_at desc").First(&build).Error
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

// SetStatus : sets the status of a build and its respective environment
func (b *Build) SetStatus(id string, status string) error {
	var err error

	tx := DB.Begin()
	tx.Exec("set transaction isolation level serializable")

	defer func() {
		switch err {
		case nil:
			err = tx.Commit().Error
		default:
			tx.Rollback()
		}
	}()

	err = tx.Raw("SELECT * FROM builds WHERE uuid = ? for update", id).Scan(b).Error
	if err != nil {
		log.Println("could not update build status")
		return err
	}

	err = tx.Exec("UPDATE builds SET status = ? WHERE id = ?", status, b.ID).Error
	if err != nil {
		log.Println("could not update build status")
		return err
	}

	err = tx.Exec("UPDATE environments SET status = ? WHERE id = ?", status, b.EnvironmentID).Error

	return err
}

// SetComponent : creates or updates a component
func (b *Build) SetComponent(c *graph.GenericComponent) error {
	return b.updateGraph(c, func(g *graph.Graph, c *graph.GenericComponent) error {
		if g.HasComponent(c.GetID()) {
			g.UpdateComponent(c)
			return nil
		}
		return g.AddComponent(c)
	})
}

// DeleteComponent : updates a component
func (b *Build) DeleteComponent(c *graph.GenericComponent) error {
	return b.updateGraph(c, func(g *graph.Graph, c *graph.GenericComponent) error {
		g.DeleteComponent(c)
		return nil
	})
}

// SetChange : updates a change
func (b *Build) SetChange(c *graph.GenericComponent) error {
	return b.updateGraph(c, func(g *graph.Graph, c *graph.GenericComponent) error {
		for i := 0; i < len(g.Changes); i++ {
			if g.Changes[i].GetID() == c.GetID() {
				g.Changes[i] = c
				return nil
			}
		}
		return errors.New("change component not found")
	})
}

// DeleteChange : deletes a change
func (b *Build) DeleteChange(c *graph.GenericComponent) error {
	return b.updateGraph(c, func(g *graph.Graph, c *graph.GenericComponent) error {
		for i := len(g.Changes) - 1; i >= 0; i-- {
			if g.Changes[i].GetID() == c.GetID() {
				g.Changes = append(g.Changes[:i], g.Changes[i+1:]...)
			}
		}
		return nil
	})
}

func (b *Build) updateGraph(c *graph.GenericComponent, tf GraphTransform) error {
	var err error

	tx := DB.Begin()
	tx.Exec("set transaction isolation level serializable")

	defer func() {
		switch err {
		case nil:
			err = tx.Commit().Error
		default:
			tx.Rollback()
		}
	}()

	err = tx.Raw("SELECT * FROM builds WHERE uuid = ? for update", (*c)["service"]).Scan(b).Error
	if err != nil {
		return err
	}

	g := graph.New()

	err = g.Load(b.Mapping)
	if err != nil {
		return err
	}

	// run graph transform function
	err = tf(g, c)
	if err != nil {
		return err
	}

	b.Mapping.LoadGraph(g)

	err = tx.Save(b).Error

	return err
}
