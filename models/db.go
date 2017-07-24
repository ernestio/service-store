/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"reflect"

	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func supported(x interface{}) []string {
	var sp []string

	rx := reflect.TypeOf(x)

	for i := 0; i < rx.NumField(); i++ {
		sp = append(sp, rx.Field(i).Tag.Get("json"))
	}

	return sp
}

func buildQuery() {

}
