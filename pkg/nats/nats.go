package nats

import (
	"errors"
	"log/slog"

	jetstreamconsumer "github.com/Mattilsynet/map-managed-environment/gen/mattilsynet/provider-jetstream-nats/jetstream-consumer"
	jetstreampublish "github.com/Mattilsynet/map-managed-environment/gen/mattilsynet/provider-jetstream-nats/jetstream-publish"
	jetstream_types "github.com/Mattilsynet/map-managed-environment/gen/mattilsynet/provider-jetstream-nats/types"
	"github.com/Mattilsynet/map-managed-environment/gen/wasmcloud/messaging/consumer"
	"github.com/Mattilsynet/map-managed-environment/gen/wasmcloud/messaging/handler"
	"github.com/Mattilsynet/map-managed-environment/gen/wasmcloud/messaging/types"
	"github.com/bytecodealliance/wasm-tools-go/cm"
)

// TODO: split this into two files, one for nats-core and one for nats-jetstream
type (
	Conn struct {
		js JetStreamContext
	}
	JetStreamContext struct {
		bucket KeyValue
	}

	Msg struct {
		Subject string
		Reply   string
		Data    []byte
		Header  map[string][]string
	}
)

type MsgHandler func(msg *Msg)

func NewConn() *Conn {
	return &Conn{}
}

func (c *Conn) Jetstream() (*JetStreamContext, error) {
	return &c.js, nil
}

func (js *JetStreamContext) Subscribe(msgHandler MsgHandler, logger *slog.Logger) {
	logger.Info("*** Subscribing to jetstream ***", "sub", "sub")
	jetstreamconsumer.Exports.HandleMessage = toWitExportSubscription(msgHandler)
}

func toWitExportSubscription(msgHandler MsgHandler) func(msg jetstreamconsumer.Msg) cm.Result[string, struct{}, string] {
	return func(msg jetstreamconsumer.Msg) cm.Result[string, struct{}, string] {
		natsMsg := &Msg{
			Subject: msg.Subject,
			Reply:   msg.Reply,
			Data:    msg.Data.Slice(),
			Header:  toNatsHeaders(msg.Headers),
		}
		msgHandler(natsMsg)
		return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
	}
}

func toNatsHeaders(header cm.List[jetstream_types.KeyValue]) map[string][]string {
	natsHeaders := make(map[string][]string)
	for _, kv := range header.Slice() {
		natsHeaders[kv.Key] = kv.Value.Slice()
	}
	return natsHeaders
}

func (js *JetStreamContext) Publish(subj string, data []byte) error {
	return js.PublishMsg(&Msg{Subject: subj, Data: data})
}

func (js *JetStreamContext) PublishMsg(msg *Msg) error {
	jpMsg := jetstreampublish.Msg{
		Headers: toWitNatsHeaders(msg.Header),
		Data:    cm.ToList(msg.Data),
		Subject: msg.Subject,
	}
	result := jetstreampublish.Publish(jpMsg)
	if !result.IsOK() {
		return errors.New(*result.Err())
	}
	return nil
}

func toWitNatsHeaders(header map[string][]string) cm.List[jetstream_types.KeyValue] {
	keyValueList := make([]jetstream_types.KeyValue, 0)
	for k, v := range header {
		keyValueList = append(keyValueList, jetstream_types.KeyValue{
			Key:   k,
			Value: cm.ToList(v),
		})
	}
	return cm.ToList(keyValueList)
}

func FromBrokerMessageToNatsMessage(bm types.BrokerMessage) *Msg {
	if bm.ReplyTo.None() {
		return &Msg{
			Data:    bm.Body.Slice(),
			Subject: bm.Subject,
			Reply:   "",
		}
	} else {
		return &Msg{
			Data:    bm.Body.Slice(),
			Subject: bm.Subject,
			Reply:   *bm.ReplyTo.Some(),
		}
	}
}

func ToBrokenMessageFromNatsMessage(nm *Msg) types.BrokerMessage {
	if nm.Reply == "" {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.None[string](),
		}
	} else {
		return types.BrokerMessage{
			Subject: nm.Subject,
			Body:    cm.ToList(nm.Data),
			ReplyTo: cm.Some(nm.Subject),
		}
	}
}

func (nc *Conn) Publish(msg *Msg) error {
	bm := ToBrokenMessageFromNatsMessage(msg)
	result := consumer.Publish(bm)
	if result.IsErr() {
		return errors.New(*result.Err())
	}
	return nil
}

func (conn *Conn) RequestReply(msg *Msg, timeoutInMillis uint32) (*Msg, error) {
	bm := ToBrokenMessageFromNatsMessage(msg)
	result := consumer.Request(bm.Subject, bm.Body, timeoutInMillis)
	if result.IsOK() {
		bmReceived := result.OK()
		natsMsgReceived := FromBrokerMessageToNatsMessage(*bmReceived)
		return natsMsgReceived, nil
	} else {
		return nil, errors.New(*result.Err())
	}
}

func (conn *Conn) RegisterRequestReply(fn func(*Msg) *Msg) {
	handler.Exports.HandleMessage = func(msg types.BrokerMessage) (result cm.Result[string, struct{}, string]) {
		natsMsg := FromBrokerMessageToNatsMessage(msg)
		newMsg := fn(natsMsg)
		return consumer.Publish(ToBrokenMessageFromNatsMessage(newMsg))
	}
}

func (conn *Conn) RegisterSubscription(fn func(*Msg)) {
	handler.Exports.HandleMessage = func(msg types.BrokerMessage) (result cm.Result[string, struct{}, string]) {
		natsMsg := FromBrokerMessageToNatsMessage(msg)
		fn(natsMsg)
		return cm.OK[cm.Result[string, struct{}, string]](struct{}{})
	}
}
