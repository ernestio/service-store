/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"

	"github.com/ernestio/service-store/models"
	"github.com/jinzhu/gorm"
)

// DepreciatedColumns : a list of columns that have been removed from the schema
var DepreciatedColumns = []string{"options", "sync", "sync_type", "sync_interval", "definition", "mapping", "endpoint", "last_known_error", "version", "user_id", "uuid", "type"}

// Migrate existing database schemas to new setup
func Migrate(db *gorm.DB) error {
	var builds []models.Build

	if db.HasTable(models.Build{}) {
		fmt.Println("has table!")
		return nil
	}

	db.CreateTable(models.Build{})
	//db.AutoMigrate(models.Service{}, models.Build{})

	// Create builds from service records
	db.Table("services").Select("id as service_id, uuid, user_id, status, mapping, definition, created_at, updated_at").Find(&builds)

	for _, b := range builds {
		// update the builds service id to the most recent service build
		var services []models.Service

		db.Table("services").Select("id, name").Where("id = ?", b.ServiceID).Find(&services)
		db.Raw("SELECT ID FROM services s1 WHERE updated_at = (SELECT MAX(updated_at) FROM services s2 WHERE s1.name = s2.name) AND name = ?;", services[0].Name).Scan(&services)

		b.ServiceID = services[0].ID
		b.Type = "apply"
		db.Table("builds").Create(&b)
	}

	// Clear out older versions of services : scary!
	db.Exec("DELETE FROM services s1 WHERE updated_at != (SELECT MAX(updated_at) FROM services s2 WHERE s1.name = s2.name);")

	// Remove options column
	for _, col := range DepreciatedColumns {
		db.Table("services").DropColumn(col)
	}

	db.AutoMigrate(models.Service{})

	// setup table relationshios
	db.Model(&models.Service{}).Related(&models.Build{})

	return nil
}
