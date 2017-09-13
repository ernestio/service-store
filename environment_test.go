/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ernestio/service-store/models"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentGet(t *testing.T) {
	cases := []struct {
		Name     string
		Query    map[string]interface{}
		Expected *models.Environment
	}{
		{"by-id", map[string]interface{}{"id": 1}, &models.Environment{ID: uint(1), Name: "Test1", Status: "done"}},
		{"by-name", map[string]interface{}{"name": "Test2"}, &models.Environment{ID: uint(2), Name: "Test2", Status: "done"}},
		{"nonexistent", map[string]interface{}{"name": "Test100"}, nil},
	}

	setupTestSuite("test_environment_get")

	db.Unscoped().Delete(models.Environment{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var e models.Environment

			data, _ := json.Marshal(tc.Query)
			resp, err := n.Request("environment.get", data, time.Second)
			assert.Nil(t, err)

			err = json.Unmarshal(resp.Data, &e)
			assert.Nil(t, err)

			if tc.Expected != nil {
				assert.Equal(t, tc.Expected.ID, e.ID)
				assert.Equal(t, tc.Expected.Name, e.Name)
				assert.Equal(t, tc.Expected.Status, e.Status)
			} else {
				assert.Equal(t, uint(0), e.ID)
				assert.Contains(t, string(resp.Data), "not found")
			}
		})
	}

}

func TestEnvironmentFind(t *testing.T) {
	cases := []struct {
		Name     string
		Query    map[string]interface{}
		Expected int
	}{
		{"by-name", map[string]interface{}{"name": "Test2"}, 1},
		{"by-status", map[string]interface{}{"status": "done"}, 20},
		{"nonexistent", map[string]interface{}{"name": "Test100"}, 0},
	}

	setupTestSuite("test_environment_get")

	db.Unscoped().Delete(models.Environment{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var es []models.Environment

			data, _ := json.Marshal(tc.Query)
			resp, err := n.Request("environment.find", data, time.Second)
			assert.Nil(t, err)

			err = json.Unmarshal(resp.Data, &es)
			assert.Nil(t, err)

			assert.Equal(t, tc.Expected, len(es))
		})
	}
}
