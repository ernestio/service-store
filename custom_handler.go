/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/ernestio/service-store/models"
	"github.com/nats-io/nats"
	graph "gopkg.in/r3labs/graph.v2"
)

// Message ...
type Message struct {
	ID         string                 `json:"id"`
	Definition string                 `json:"definition"`
	Mapping    map[string]interface{} `json:"mapping"`
}

func complete(reply string, data *[]byte, err *error) {
	if *err != nil {
		log.Println(*err)
		d := []byte(`{"error": "` + (*err).Error() + `"}`)
		data = &d
	} else if data == nil {
		d := []byte(`{"status": "success"}`)
		data = &d
	}

	if reply != "" {
		handler.Nats.Publish(reply, *data)
	}
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
	var m *Message
	var b *models.Build
	var err error
	var data []byte

	defer complete(msg.Reply, &data, &err)

	m, err = getMessage(msg)
	if err != nil {
		return
	}

	b, err = models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		return
	}

	if b == nil {
		err = errors.New("build not found")
		return
	}

	data, err = json.Marshal(b.Mapping)
}

// SetMapping : Mapping field setter
func SetMapping(msg *nats.Msg) {
	var m *Message
	var b *models.Build
	var err error

	defer complete(msg.Reply, nil, &err)

	m, err = getMessage(msg)
	if err != nil {
		return
	}

	b, err = models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		return
	}

	b.Mapping = m.Mapping

	err = b.Update()
}

// GetDefinition : Definition field getter
func GetDefinition(msg *nats.Msg) {
	var m *Message
	var b *models.Build
	var err error
	var data []byte

	defer complete(msg.Reply, &data, &err)

	m, err = getMessage(msg)
	if err != nil {
		return
	}

	b, err = models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		return
	}

	data = []byte(b.Definition)
}

// SetDefinition : Definition field setter
func SetDefinition(msg *nats.Msg) {
	var m *Message
	var b *models.Build
	var err error

	defer complete(msg.Reply, nil, &err)

	m, err = getMessage(msg)
	if err != nil {
		return
	}

	b, err = models.GetBuild(map[string]interface{}{"uuid": m.ID})
	if err != nil {
		return
	}

	b.Definition = m.Definition

	err = b.Update()
}

// SetComponent : Mapping component setter
func SetComponent(msg *nats.Msg) {
	var c *graph.GenericComponent
	var b models.Build
	var err error

	defer complete(msg.Reply, nil, &err)

	c, err = getComponent(msg)
	if err != nil {
		return
	}

	err = b.SetComponent(c)
}

// DeleteComponent : Mapping component deleter
func DeleteComponent(msg *nats.Msg) {
	var c *graph.GenericComponent
	var b models.Build
	var err error

	defer complete(msg.Reply, nil, &err)

	c, err = getComponent(msg)
	if err != nil {
		return
	}

	err = b.DeleteComponent(c)
}

// SetChange : Mapping change setter
func SetChange(msg *nats.Msg) {
	var c *graph.GenericComponent
	var b models.Build
	var err error

	defer complete(msg.Reply, nil, &err)

	c, err = getComponent(msg)
	if err != nil {
		return
	}

	err = b.SetChange(c)
}

// ServiceComplete : sets a services error to complete
func ServiceComplete(msg *nats.Msg) {
	var b models.Build
	var err error

	parts := strings.Split(msg.Subject, ".")

	m, err := getMessage(msg)
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}

	err = b.SetStatus(m.ID, "done")
	if err != nil {
		log.Println("could not handle service complete message: " + err.Error())
	}

	if parts[1] == "delete" {
		b, err := models.GetBuild(map[string]interface{}{"uuid": m.ID})
		if err != nil {
			log.Println("could not get build from service complete message: " + err.Error())
		}

		s, err := models.GetService(map[string]interface{}{"id": b.ServiceID})
		if err != nil {
			log.Println("could not get service from service complete message: " + err.Error())
		}

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
