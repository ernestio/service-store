/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
	"github.com/r3labs/graph"
)

type Message struct {
	ID         string                 `json:"id"`
	Definition string                 `json:"definition"`
	Mapping    map[string]interface{} `json:"mapping"`
}

func getMessage(msg *nats.Msg) (*Message, error) {
	var m Message
	return &m, json.Unmarshal(msg.Data, &m)
}

func getComponent(msg *nats.Msg) (*graph.GenericComponent, error) {
	var c graph.GenericComponent
	return &c, json.Unmarshal(msg.Data, &c)
}

// GetMapping : Mapping field getter
func GetMapping(msg *nats.Msg) {
	m, err := getMessage(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil || b == nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	data, err := json.Marshal(b.Mapping)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	handler.Nats.Publish(msg.Reply, data)
}

// SetMapping : Mapping field setter
func SetMapping(msg *nats.Msg) {
	m, err := getMessage(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "could not find build"}`))
	}

	b.Mapping = m.Mapping

	err = b.Update()
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "could not store build mapping"}`))
	}

	handler.Nats.Publish(msg.Reply, []byte(`{"status": "success"}`))
}

// GetDefinition : Definition field getter
func GetDefinition(msg *nats.Msg) {
	m, err := getMessage(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	handler.Nats.Publish(msg.Reply, []byte(b.Definition))
}

// SetDefinition : Definition field setter
func SetDefinition(msg *nats.Msg) {
	m, err := getMessage(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	b.Definition = m.Definition

	err = b.Update()
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	handler.Nats.Publish(msg.Reply, []byte(`{"status": "success"}`))
}

// SetComponent : Mapping component setter
func SetComponent(msg *nats.Msg) {
	var b models.Build

	c, err := getComponent(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	err = tx.Raw("SELECT * FROM builds WHERE uuid = ? for update", (*c)["service"]).Scan(&b).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = b.SetComponent(c)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	err = tx.Save(&b).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	_ = handler.Nats.Publish(msg.Reply, []byte(`{"status":"success"}`))
}

// DeleteComponent : Mapping component deleter
func DeleteComponent(msg *nats.Msg) {
	var b models.Build

	c, err := getComponent(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	err = tx.Raw("SELECT * FROM builds WHERE uuid = ? for update", (*c)["service"]).Scan(&b).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = b.DeleteComponent(c)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	err = tx.Save(&b).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	_ = handler.Nats.Publish(msg.Reply, []byte(`{"status":"success"}`))
}

// SetChange : Mapping change setter
func SetChange(msg *nats.Msg) {
	var b models.Build

	c, err := getComponent(msg)
	if err != nil {
		handler.Nats.Publish(msg.Reply, []byte(`{"error": "bad request"}`))
	}

	tx := db.Begin()
	tx.Exec("set transaction isolation level serializable")

	err = tx.Raw("SELECT * FROM builds WHERE uuid = ? for update", (*c)["service"]).Scan(&b).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = b.SetChange(c)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	err = tx.Save(&b).Error
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return
	}

	tx.Commit()

	_ = handler.Nats.Publish(msg.Reply, []byte(`{"status":"success"}`))
}

// ServiceDeleteComplete : sets a services error to complete
func ServiceDeleteComplete(msg *nats.Msg) {
	parts := strings.Split(msg.Subject, ".")

	m, err := getMessage(msg)
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		log.Println("could not get build from service complete message: " + err.Error())
	}

	s, err := models.GetService(map[string]interface{}{"id": b.ServiceID})
	if err != nil {
		log.Println("could not get service from service complete message: " + err.Error())
	}

	if parts[1] == "delete" {
		err = s.Delete()
		if err != nil {
			log.Println("could not get delete the service: " + err.Error())
		}
	}
}

// ServiceError : sets a services error to errored
func ServiceError(msg *nats.Msg) {
	m, err := getMessage(msg)
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}

	b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		log.Println("could not get build from service complete message: " + err.Error())
	}

	s, err := models.GetService(map[string]interface{}{"id": b.ServiceID})
	if err != nil {
		log.Println("could not get service from service complete message: " + err.Error())
	}

	s.Status = "errored"
	b.Status = "errored"

	err = b.Update()
	if err != nil {
		log.Println("could not save build from service error message: " + err.Error())
	}

	err = s.Update()
	if err != nil {
		log.Println("could not save service from service error message: " + err.Error())
	}
}
