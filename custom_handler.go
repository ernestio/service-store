package main

import (
	"github.com/nats-io/nats"
)

// GetMapping : Mapping field getter
func GetMapping(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		_ = handler.Nats.Publish(msg.Reply, []byte(e.Mapping))
	}
}

// SetMapping : Mapping field setter
func SetMapping(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		input := Entity{}
		input.MapInput(msg.Data)
		e.Mapping = input.Mapping
		db.Save(&e)
		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}

// GetDefinition : Definition field getter
func GetDefinition(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		_ = handler.Nats.Publish(msg.Reply, []byte(e.Definition))
	}
}

// SetDefinition : Definition field setter
func SetDefinition(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		input := Entity{}
		input.MapInput(msg.Data)
		e.Definition = input.Definition
		db.Save(&e)
		_ = handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}
