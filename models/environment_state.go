/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/r3labs/statemachine"
)

// StatePayload : stores informmation about the current triggered event
type StatePayload struct {
	EnvironmentID uint
	tx            *gorm.DB
}

var (
	BaseEvents    = []string{"initializing", "done", "errored"}
	SpecialEvents = []string{"sync-accepted", "sync-ignored", "sync-rejected", "submission-accepted", "submission-rejected"}
)

// NewStateMachine ...
func NewStateMachine(e *Environment) *statemachine.StateMachine {
	sm := statemachine.New(e)

	sm.When("apply", statemachine.Transitions{"initializing": "in_progress", "done": "in_progress", "errored": "in_progress"})
	sm.When("sync", statemachine.Transitions{"initializing": "syncing", "done": "syncing", "errored": "syncing"})
	sm.When("submission", statemachine.Transitions{"initializing": "awaiting_approval", "done": "awaiting_approval", "errored": "awaiting_approval"})
	sm.When("sync-rejected", statemachine.Transitions{"awaiting_resolution": "in_progress"})
	sm.When("sync-accepted", statemachine.Transitions{"awaiting_resolution": "done"})
	sm.When("sync-ignored", statemachine.Transitions{"awaiting_resolution": "done"})
	sm.When("submission-accepted", statemachine.Transitions{"awaiting_approval": "in_progress"})
	sm.When("submission-rejected", statemachine.Transitions{"awaiting_approval": "done"})

	sm.Error("syncing", errors.New("could not create environment build: environment is syncing"))
	sm.Error("in_progress", errors.New("could not create environment build: environment in progress"))

	for _, e := range SpecialEvents {
		sm.On(e, CallbackLastBuildStatus)
	}

	for _, e := range append(BaseEvents, SpecialEvents...) {
		sm.On(e, CallbackEnvironmentStatus)
	}

	return sm
}

// CallbackLastBuildStatus : sets the last build status
func CallbackLastBuildStatus(state string, p interface{}) error {
	sp, ok := p.(*StatePayload)
	if !ok {
		return errors.New("unknown state payload")
	}

	return SetLatestBuildStatus(sp.EnvironmentID, "done")
}

// CallbackEnvironmentStatus : sets the environments status
func CallbackEnvironmentStatus(state string, p interface{}) error {
	sp, ok := p.(*StatePayload)
	if !ok {
		return errors.New("unknown state payload")
	}

	return sp.tx.Exec("UPDATE environments SET status = ? WHERE id = ?", state, sp.EnvironmentID).Error
}
