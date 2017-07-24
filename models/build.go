/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"time"
)

// Build : stores build data
type Build struct {
	ID         uint       `json:"-" gorm:"primary_key"`
	UUID       string     `json:"id"`
	ServiceID  uint       `json:"service_id"`
	UserID     uint       `json:"user_id"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	Definition string     `json:"definition" gorm:"type:text;"`
	Mapping    Mapping    `type: jsonb not null default '{}'::jsonb`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"-" sql:"index"`
}
