//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world map-managed-environment --out gen ./wit
package main

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/Mattilsynet/map-managed-environment/pkg/nats"
	"github.com/Mattilsynet/map-managed-environment/pkg/subject"
	"github.com/Mattilsynet/mapis/gen/go/command/v1"
	"github.com/Mattilsynet/mapis/gen/go/managedenvironment/v1"
	"github.com/Mattilsynet/mapis/gen/go/query/v1"
	"go.wasmcloud.dev/component/log/wasilog"
)

var (
	conn   *nats.Conn
	js     *nats.JetStreamContext
	logger *slog.Logger
	kv     *nats.KeyValue
	re     *regexp.Regexp
)

const resultSubjectTemplate = "map.result.%s.%s.ManagedEnvironment"

func init() {
	logger = wasilog.ContextLogger("map-managed-environment")
	conn = nats.NewConn()
	var err error
	js, err = conn.Jetstream()
	if err != nil {
		logger.Error("error getting jetstreamcontext", "err", err)
		return
	}
	kv, err = js.KeyValue()
	if kv == nil {
		logger.Error("error getting keyvalue", "err", err)
		return
	}
	logger.Info("*** map-managed-environment started ***", "sub", "pub-dub")
	js.Subscribe(ManagedEnvironmentConsumer, logger)
}

/*
TODO:
extract:
1. queryHandler to a seperate file
2. commandHandler to a seperate file
3. make common used strings as constants
4. make error handling more generic
*/
func ManagedEnvironmentConsumer(msg *nats.Msg) {
	crudOperation := subject.GetLastPartition(msg.Subject)
	lowerCasedCrudOperation := strings.ToLower(crudOperation)
	logger.Info("crudOperation", "crudOperation", lowerCasedCrudOperation)
	switch lowerCasedCrudOperation {
	case "get":
		queryHandler(msg)
	default:
		commandHandler(msg)
	}
}

func queryHandler(msg *nats.Msg) {
	resultMsg := &nats.Msg{}
	qry := &query.Query{}
	err := qry.UnmarshalVT(msg.Data)
	if err != nil {
		logger.Info("query unmarshalVT failed with err", "err", err)
		resultMsg.Data = []byte("query unmarshalVT failed with err: " + err.Error())
		resultMsg.Subject = fmt.Sprintf(resultSubjectTemplate, "unknown", "unknown")
		conn.Publish(resultMsg)
		return
	}
	logger.Info("1:")
	resultSubject := fmt.Sprintf(resultSubjectTemplate, qry.Spec.Session, qry.Status.Id)
	resultMsg.Subject = resultSubject
	if qry.Spec.Type.Kind != "ManagedEnvironment" {
		logger.Info("2: wrong me")
		resultMsg.Data = []byte("query.specc.Type.Kind is not ManagedEnvironment, got: " + qry.Spec.Type.Kind)
		conn.Publish(resultMsg)
		return
	}
	queryFilter := qry.Spec.QueryFilter
	// INFO: we fetch all managed-environments in KV
	if queryFilter.Name == "" {
		logger.Info("inside no queryFilter.Name")
		if queryFilter.StatusFilter == nil {
			kves, err := kv.GetAll()
			if err != nil {
				resultMsg.Data = []byte("ManagedEnvironment get all failed with: " + err.Error())
				conn.Publish(resultMsg)
				return
			}
			managedEnvironmentsAsBytes := make([][]byte, len(kves))
			for i, kve := range kves {
				managedEnvironmentsAsBytes[i] = kve.Value()
			}
			qry.Status.TypePayload = managedEnvironmentsAsBytes // info: map each data element of kves to [][]byte
			// marshal qry to bytes
			qryData, err := qry.MarshalVT()
			if err != nil {
				resultMsg.Data = []byte("marshalVT on MES and not from meExisting failed with err: " + err.Error() + " withour ManagedEnvironment filter " + "please contact Plattform team, don't try again")
				conn.Publish(resultMsg)
				return
			}
			resultMsg.Data = qryData
			conn.Publish(resultMsg)
			return
		} else {
			return
			// INFO: handle statusFilter
		}
	} else {
		logger.Info("inside queryFilter.Name")
		if queryFilter.StatusFilter == nil {
			logger.Info("inside queryFilter.Name, no StatusFilter")
			kve, err := kv.Get(queryFilter.Name)
			if err != nil {
				if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "deleted") {
					resultMsg.Data = []byte("ManagedEnvironment " + queryFilter.Name + " not found")
					conn.Publish(resultMsg)
					return

				}
				resultMsg.Data = []byte("kv.Get with given ManagedEnvironemnt: " + queryFilter.Name + " failed with err: " + err.Error() + " please contact Plattform team, don't try again")
				conn.Publish(resultMsg)
				return
			}
			bytess := make([][]byte, 0)
			qry.Status.TypePayload = append(bytess, kve.Value())
			// marshal qry to bytes
			qryData, err := qry.MarshalVT()
			if err != nil {
				resultMsg.Data = []byte("marshalVT on ME and not from meExisting failed with err: " + err.Error() + " withour ManagedEnvironment filter " + "please contact Plattform team, don't try again")
				conn.Publish(resultMsg)
				return
			}
			resultMsg.Data = qryData
			conn.Publish(resultMsg)
			return

		} else {
			logger.Info("inside queryFilter.Name, has StatusFilter")
			// INFO: handle statusFilter
		}
	}
}

/*
INFO:
1. Unmarshal command
2. Unmarshal managed-environment
3. Do a get by managed-environment name towards KV
4. Copy status from kv to this managed-environment
5. Store updated managed-environment to kv
6. publish result by command sessionid, id towards result subject
*/
func commandHandler(msg *nats.Msg) {
	resultMsg := &nats.Msg{}
	cmd := &command.Command{}
	err := cmd.UnmarshalVT(msg.Data)
	if err != nil {
		logger.Info("command unmarshalVT failed with err", "err", err)
		resultMsg.Data = []byte("command unmarshalVT failed with err: " + err.Error())
		resultMsg.Subject = fmt.Sprintf(resultSubjectTemplate, "unknown", "unknown")
		conn.Publish(resultMsg)
		return
	}
	logger.Info("1")
	resultSubject := fmt.Sprintf(resultSubjectTemplate, cmd.Spec.SessionId, cmd.Status.Id)
	resultMsg.Subject = resultSubject
	me := &managedenvironment.ManagedEnvironment{}
	if cmd.Spec.Type.Kind != "ManagedEnvironment" {
		logger.Info("2")
		resultMsg.Data = []byte("command.specc.Type.Kind is not ManagedEnvironment, got: " + cmd.Spec.Type.Kind)
		conn.Publish(resultMsg)
		return
	}
	err = me.UnmarshalVT(cmd.Spec.TypePayload)
	logger.Info("3")
	if err != nil {
		logger.Info("4")
		resultMsg.Data = []byte("command.spec.typepayload is not ManagedEnvironment with unmarshalVT err: " + err.Error())
		conn.Publish(resultMsg)
		return
	}
	cmdOperation := cmd.GetSpec().Operation
	meKey := me.Metadata.Name
	switch cmdOperation {
	case "apply":
		kve, err := kv.Get(meKey)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "deleted") {
				err = kv.Create(meKey, cmd.Spec.TypePayload)
				if err != nil {
					resultMsg.Data = []byte("kv.Create failed with err: " + err.Error() + " with ManagedEnvironment: " + meKey)
					conn.Publish(resultMsg)
					return
				}
				resultMsg.Data = []byte("ManagedEnvironment " + meKey + " created")
				conn.Publish(resultMsg)
				return

			}
			resultMsg.Data = []byte("kv.Get with given ManagedEnvironemnt: " + meKey + "failed with err: " + err.Error() + " please contact Plattform team, don't try again")
			conn.Publish(resultMsg)
			return
		}
		meExistingBytes := kve.Value()
		meExisting := &managedenvironment.ManagedEnvironment{}
		err = meExisting.UnmarshalVT(meExistingBytes)
		if err != nil {
			resultMsg.Data = []byte("unmarshalVT failed with err: " + err.Error() + " with ManagedEnvironment: " + meKey)
			conn.Publish(resultMsg)
			return
		}
		// TODO: When we updated map-type to include managedenvironment status, update me.status with meExisting status then do a kv.put(me)
		me.Status = meExisting.Status
		var bytes []byte
		bytes, err = me.MarshalVT()
		if err != nil {
			resultMsg.Data = []byte("marshalVT on ME and not from meExisting failed with err: " + err.Error() + " with ManagedEnvironment: " + meKey + " please contact Plattform team, don't try again")
			conn.Publish(resultMsg)
			return
		}
		err = kv.Put(meKey, bytes)
		if err != nil {
			resultMsg.Data = []byte("kv.Put failed with err: " + err.Error() + " with ManagedEnvironment: " + meKey)
			conn.Publish(resultMsg)
			return
		}
		resultMsg.Data = []byte("ManagedEnvironment " + meKey + " updated")
		conn.Publish(resultMsg)
	case "delete":
		err := kv.Delete(meKey)
		if err != nil {
			resultMsg.Data = []byte("kv.Delete failed with err: " + err.Error())
			conn.Publish(resultMsg)
			return
		}
		resultMsg.Data = []byte("ManagedEnvironment " + meKey + " deleted")
		conn.Publish(resultMsg)
	default:
		logger.Info("9")
		resultMsg.Data = []byte("command.spec.operation is not apply or delete, got: " + cmdOperation)
		conn.Publish(resultMsg)
		return
	}
}

//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world starter-kit --out gen ./wit
func main() {}
