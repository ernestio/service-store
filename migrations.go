/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"log"

	"github.com/ernestio/service-store/models"
	"github.com/jinzhu/gorm"
)

// DepreciatedColumns : a list of columns that have been removed from the schema
var DepreciatedColumns = []string{"options", "sync", "sync_type", "sync_interval", "definition", "mapping", "last_known_error", "version", "user_id", "uuid", "type"}

// Migrate existing database schemas to new setup
func Migrate(db *gorm.DB) error {
	// var builds []models.Build
	if db.HasTable(models.Environment{}) {
		err := db.Exec("ALTER TABLE environments RENAME COLUMN datacenter_id TO project_id;").Error
		if err != nil {
			log.Println(err)
		}
	}

	return db.AutoMigrate(models.Environment{}, models.Build{}).Error

	/*

		db.CreateTable(models.Build{})

		// Create builds from service records
		db.Table("environments").Select("id as service_id, uuid, user_id, status, mapping, definition, created_at, updated_at").Find(&builds)

		for _, b := range builds {
			// update the builds service id to the most recent service build
			var environments []models.Service

			db.Table("environments").Select("id, name").Where("id = ?", b.EnvironmentID).Find(&environments)
			db.Raw("SELECT ID FROM environments s1 WHERE updated_at = (SELECT MAX(updated_at) FROM environments s2 WHERE s1.name = s2.name) AND name = ?;", environments[0].Name).Scan(&environments)

			b.EnvironmentID = environments[0].ID
			b.Type = "apply"
			db.Table("builds").Create(&b)
		}

		// Clear out older versions of environments : scary!
		db.Exec("DELETE FROM environments s1 WHERE updated_at != (SELECT MAX(updated_at) FROM environments s2 WHERE s1.name = s2.name);")

		// Remove options column
		for _, col := range DepreciatedColumns {
			db.Table("environments").DropColumn(col)
		}

		db.AutoMigrate(models.Service{})
	*/
}
