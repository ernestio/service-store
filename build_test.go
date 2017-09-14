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

func TestBuildGet(t *testing.T) {
	cases := []struct {
		Name     string
		Query    map[string]interface{}
		Expected *models.Build
	}{
		{"by-id", map[string]interface{}{"id": "uuid-3"}, &models.Build{UUID: "uuid-3", EnvironmentID: uint(3), UserID: uint(3), Status: "done"}},
		{"nonexistent", map[string]interface{}{"id": "uuid-10000"}, nil},
	}

	setupTestSuite("test_build_get")

	db.Unscoped().Delete(models.Build{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var e models.Build

			data, _ := json.Marshal(tc.Query)
			resp, err := n.Request("build.get", data, time.Second)
			assert.Nil(t, err)

			err = json.Unmarshal(resp.Data, &e)
			assert.Nil(t, err)

			if tc.Expected != nil {
				assert.Equal(t, tc.Expected.UUID, e.UUID)
				assert.Equal(t, tc.Expected.EnvironmentID, e.EnvironmentID)
				assert.Equal(t, tc.Expected.UserID, e.UserID)
				assert.Equal(t, tc.Expected.Status, e.Status)
			} else {
				assert.Equal(t, uint(0), e.ID)
				assert.Contains(t, string(resp.Data), "not found")
			}
		})
	}

}

func TestBuildFind(t *testing.T) {
	cases := []struct {
		Name     string
		Query    map[string]interface{}
		Expected int
	}{
		{"by-id", map[string]interface{}{"id": "uuid-1"}, 1},
		{"by-status", map[string]interface{}{"status": "done"}, 20},
		{"by-environment-id", map[string]interface{}{"environment_id": 2}, 1},
		{"nonexistent", map[string]interface{}{"id": "uuid-10000"}, 0},
	}

	setupTestSuite("test_build_find")

	db.Unscoped().Delete(models.Build{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var es []models.Build

			data, _ := json.Marshal(tc.Query)
			resp, err := n.Request("build.find", data, time.Second)
			assert.Nil(t, err)

			err = json.Unmarshal(resp.Data, &es)
			assert.Nil(t, err)

			assert.Equal(t, tc.Expected, len(es))
		})
	}
}

func TestBuildSet(t *testing.T) {
	cases := []struct {
		Name     string
		Event    *models.Build
		Expected *models.Build
	}{
		{"existing", &models.Build{UUID: "uuid-1", EnvironmentID: uint(1), Status: "done"}, &models.Build{UUID: "uuid-1", Status: "done"}},
		{"nonexistent", &models.Build{EnvironmentID: uint(2), Status: "in_progress"}, &models.Build{UUID: "GENERATED", Status: "in_progress"}},
	}

	setupTestSuite("test_build_set")

	db.Unscoped().Delete(models.Build{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var e models.Build

			data, _ := json.Marshal(tc.Event)
			resp, err := n.Request("build.set", data, time.Second)
			assert.Nil(t, err)

			err = json.Unmarshal(resp.Data, &e)
			assert.Nil(t, err)

			if tc.Expected.UUID == "GENERATED" {
				assert.NotEqual(t, tc.Expected.UUID, "")
			} else {
				assert.Equal(t, tc.Expected.UUID, e.UUID)
			}
			assert.Equal(t, tc.Expected.Status, e.Status)
		})
	}
}

func TestBuildDelete(t *testing.T) {
	cases := []struct {
		Name     string
		Event    *models.Build
		Expected string
	}{
		{"existing", &models.Build{UUID: "uuid-1"}, "success"},
	}

	setupTestSuite("test_build_delete")

	db.Unscoped().Delete(models.Build{}, models.Build{})
	CreateTestData(db, 20)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			data, _ := json.Marshal(tc.Event)
			resp, err := n.Request("build.del", data, time.Second)
			assert.Nil(t, err)

			assert.Contains(t, string(resp.Data), tc.Expected)
		})
	}
}
