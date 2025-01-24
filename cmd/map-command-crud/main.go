package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Mattilsynet/mapis/gen/go/command/v1"
	metav1 "github.com/Mattilsynet/mapis/gen/go/meta/v1"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Error connecting to nats: %v", err)
	}
	cmd := command.Command{}
	cmd.Spec = &command.CommandSpec{Operation: "create", SessionId: "123"}
	cmd.Spec.Type = &metav1.TypeMeta{Kind: "idporten", ApiVersion: "v1"}
	cmd.Spec.TypePayload = []byte("stuff")
	bytes, err := json.Marshal(&cmd)
	if err != nil {
		log.Fatalf("Error marshalling command: %v", err)
	}
	res, err := nc.Request("map.create", bytes, 10*time.Second)
	if err != nil {
		log.Fatalf("Error sending command: %v", err)
	}
	log.Println("response: ", string(res.Data))
}
