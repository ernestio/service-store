/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"strconv"
)

func setupTestSuite() {
	db.Where("true = true").Unscoped().Delete(Entity{})
}

func createEntities(n int) {
	i := 0
	for i < n {
		x := strconv.Itoa(i)
		db.Create(&Entity{
			Name:         "Test" + x,
			Uuid:         "random_string" + x,
			GroupID:      1,
			DatacenterID: 1,
			Type:         "type",
			Version:      "0.0.1",
			Status:       "in_progress",
			Options:      "options",
			Definition:   "definition",
			Mapping:      `{"valid":"json"}`,
		})
		i++
	}
}
