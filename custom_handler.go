package main

import (
	"github.com/nats-io/nats"
)

func GetMapping(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		handler.Nats.Publish(msg.Reply, []byte(e.Mapping))
	}
}

func SetMapping(msg *nats.Msg) {
	e := Entity{}
	if ok := e.LoadFromInputOrFail(msg, &handler); ok {
		input := Entity{}
		input.MapInput(msg.Data)
		e.Mapping = input.Mapping
		db.Save(&e)
		handler.Nats.Publish(msg.Reply, []byte(`"success"`))
	}
}
