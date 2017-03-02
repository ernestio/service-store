/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"strconv"
	"time"
)

func setupTestSuite() {
	db.Unscoped().Delete(Entity{})
}

func createEntities(n int) {
	i := 0
	for i < n {
		x := strconv.Itoa(i)
		db.Create(&Entity{
			Name:         "Test" + x,
			UUID:         "random_string" + x,
			GroupID:      1,
			DatacenterID: 1,
			Type:         "type",
			Version:      time.Now(),
			Status:       "in_progress",
			Options:      "options",
			Definition:   "definition",
			Mapping:      `{"valid":"json","endpoint":"1.1.1.1"}`,
		})
		i++
	}
}
