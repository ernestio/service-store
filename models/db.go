/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func structFields(x interface{}) []string {
	var sp []string

	rx := reflect.TypeOf(x)

	for i := 0; i < rx.NumField(); i++ {
		sp = append(sp, rx.Field(i).Tag.Get("json"))
	}

	return sp
}

func supported(t interface{}, f string) bool {
	for _, field := range structFields(t) {
		if f == field {
			return true
		}
	}
	return false
}

func query(q map[string]interface{}, results interface{}) {
	qdb := DB

	t := reflect.TypeOf(results).Elem().Kind()

	for k, v := range q {
		if supported(t, k) {
			qdb = qdb.Where("? = ?", k, v)
		}
	}

	qdb.Find(results)
}
