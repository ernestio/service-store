package main

import (
	"log"
	"strings"

	"github.com/nats-io/nats"
)

// GetMapping : Mapping field getter
func GetMapping(msg *nats.Msg) {
	b := Build{}
	if ok := b.LoadFromInputOrFail(msg, &handler); ok {
		_ = handler.Nats.Publish(msg.Reply, []byte(b.Mapping))
	}
}

// SetMapping : Mapping field setter
func SetMapping(msg *nats.Msg) {
	b := Build{}
	if ok := s.LoadFromInputOrFail(msg, &handler); ok {
		input := Build{}
		input.MapInput(msg.Data)
		b.Mapping = input.Mapping
		db.Save(&e)
		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// GetDefinition : Definition field getter
func GetDefinition(msg *nats.Msg) {
	b := Build{}
	if ok := s.LoadFromInputOrFail(msg, &handler); ok {
		_ = handler.Nats.Publish(msg.Reply, []byte(e.Definition))
	}
}

// SetDefinition : Definition field setter
func SetDefinition(msg *nats.Msg) {
	b := Build{}
	if ok := s.LoadFromInputOrFail(msg, &handler); ok {
		input := Build{}
		input.MapInput(msg.Data)
		b.Definition = input.Definition
		db.Save(&b)
		_ = handler.Nats.Publish(msg.Reply, []byte(`{"status": "success"}`))
	}
}

// SetComponent : Mapping component setter
func SetComponent(msg *nats.Msg) {
	var c Component
	s := Service{}
	if ok := c.LoadFromInputOrFail(msg, &handler); ok {
		sid, _ := c.GetServiceID()

		tx := db.Begin()
		tx.Exec("set transaction isolation level serializable")

		err := tx.Raw("SELECT * FROM services WHERE uuid = ? for update", sid).Scan(&s).Error
		if err != nil {
			tx.Rollback()
			return
		}

		err = s.setComponent(c)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		err = tx.Save(&s).Error
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		tx.Commit()

		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// DeleteComponent : Mapping component deleter
func DeleteComponent(msg *nats.Msg) {
	var c Component
	s := Service{}
	if ok := c.LoadFromInputOrFail(msg, &handler); ok {
		sid, _ := c.GetServiceID()
		cid, _ := c.GetComponentID()

		tx := db.Begin()
		tx.Exec("set transaction isolation level serializable")

		tx.Where("uuid = ?", sid).First(&s)
		if &s == nil {
			tx.Rollback()
			return
		}

		err := s.deleteComponent(cid)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		err = tx.Save(&s).Error
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		tx.Commit()

		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// SetChange : Mapping change setter
func SetChange(msg *nats.Msg) {
	var c Component
	s := Service{}
	if ok := c.LoadFromInputOrFail(msg, &handler); ok {
		sid, _ := c.GetServiceID()

		tx := db.Begin()
		tx.Exec("set transaction isolation level serializable")

		err := tx.Raw("SELECT * FROM services WHERE uuid = ? for update", sid).Scan(&s).Error
		if err != nil {
			log.Println("could not find service! " + sid)
			tx.Rollback()
			return
		}

		err = s.setChange(c)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		err = tx.Save(&s).Error
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		tx.Commit()

		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// DeleteChange : Mapping change deleter
func DeleteChange(msg *nats.Msg) {
	var c Component
	s := Service{}
	if ok := c.LoadFromInputOrFail(msg, &handler); ok {
		sid, _ := c.GetServiceID()
		cid, _ := c.GetComponentID()

		tx := db.Begin()
		tx.Exec("set transaction isolation level serializable")

		tx.Where("uuid = ?", sid).First(&s)
		if &s == nil {
			tx.Rollback()
			return
		}

		err := s.deleteChange(cid)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		err = tx.Save(&s).Error
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}

		tx.Commit()

		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// ServiceComplete : sets a services status to complete
func ServiceComplete(msg *nats.Msg) {
	parts := strings.Split(msg.Subject, ".")

	s := Service{}
	if ok := s.LoadFromInputOrFail(msg, &handler); ok {
		input := Service{}
		input.MapInput(msg.Data)

		if parts[1] == "delete" {
			_ = e.Delete()
		} else {
			if e.Status != "syncing" {
				e.Status = "done"
			}
			db.Save(&e)
		}
	}
}

// ServiceError : sets a services status to errored
func ServiceError(msg *nats.Msg) {
	s := Service{}
	if ok := s.LoadFromInputOrFail(msg, &handler); ok {
		input := Service{}
		input.MapInput(msg.Data)

		e.Status = "errored"
		db.Save(&e)
	}
}
