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
	Action        string
	tx            *gorm.DB
}

var (
	BaseStates = []string{"initializing", "done", "errored"}
)

// NewStateMachine ...
func NewStateMachine(e *Environment) *statemachine.StateMachine {
	sm := statemachine.New(e)

	sm.When("apply", statemachine.Transitions{"initializing": "in_progress", "done": "in_progress", "errored": "in_progress"})
	sm.When("destroy", statemachine.Transitions{"initializing": "in_progress", "done": "in_progress", "errored": "in_progress"})
	sm.When("import", statemachine.Transitions{"initializing": "in_progress", "done": "in_progress", "errored": "in_progress"})
	sm.When("sync", statemachine.Transitions{"initializing": "syncing", "done": "syncing", "errored": "syncing"})
	sm.When("submission", statemachine.Transitions{"initializing": "awaiting_approval", "done": "awaiting_approval", "errored": "awaiting_approval"})
	sm.When("sync-rejected", statemachine.Transitions{"awaiting_resolution": "in_progress"})
	sm.When("sync-accepted", statemachine.Transitions{"awaiting_resolution": "done"})
	sm.When("sync-ignored", statemachine.Transitions{"awaiting_resolution": "done"})
	sm.When("submission-accepted", statemachine.Transitions{"awaiting_approval": "in_progress"})
	sm.When("submission-rejected", statemachine.Transitions{"awaiting_approval": "done"})

	sm.Error("syncing", errors.New("could not create environment build: environment is syncing"))
	sm.Error("in_progress", errors.New("could not create environment build: build in progress"))

	for _, e := range BaseStates {
		sm.On(e, CallbackUpdateStatus)
	}

	return sm
}

// CallbackUpdateStatus : sets the environments status
func CallbackUpdateStatus(state string, p interface{}) error {
	var err error

	sp, ok := p.(*StatePayload)
	if !ok {
		return errors.New("unknown state payload")
	}

	switch sp.Action {
	case "sync-accepted", "sync-ignored", "sync-rejected":
		err = SetLatestBuildStatus(sp.EnvironmentID, "done")
	}

	if err != nil {
		return err
	}

	return sp.tx.Exec("UPDATE environments SET status = ? WHERE id = ?", state, sp.EnvironmentID).Error
}
