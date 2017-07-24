/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

// Migrate existing database schemas to new setup
func Migrate(db *gorm.DB) error {
	if db.HasTable() {

	}
	return nil
}

func tableExists(db *sql.DB, name string) bool {
	return false
}
