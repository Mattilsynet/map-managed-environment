package subject

import "github.com/Mattilsynet/mapis/gen/go/command/v1"

const (
	CREATE Operation = "create"
	UPDATE Operation = "update"
	DELETE Operation = "delete"
	GET    Operation = "get"
)

type Operation string

type CommandQuerySubject struct {
	Kind    string
	Id      string
	Session string
	prefix  string
}

func NewCommandQuerySubject(command *command.Command) *CommandQuerySubject {
	id := command.GetStatus().GetId()
	session := command.Spec.GetSessionId()
	kind := command.GetSpec().GetType().GetKind()
	prefix := "map"
	return &CommandQuerySubject{
		Kind:    kind,
		Session: session,
		Id:      id,
		prefix:  prefix,
	}
}

func (qs *CommandQuerySubject) ToCommand(command *command.Command) string {
	operation := command.GetSpec().GetOperation()
	subject := "map" + "." + qs.Kind + "." + qs.Session + "." + qs.Id + "." + operation
	return subject
}

func (qs *CommandQuerySubject) ToQuery(operation Operation) string {
	subject := "map" + "." + qs.Kind + "." + qs.Session + "." + qs.Id + "." + string(operation)
	return subject
}
