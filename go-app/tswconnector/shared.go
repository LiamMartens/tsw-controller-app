package tswconnector

import (
	"fmt"
	"sort"
	"strings"
)

type TSWConnector_Message_EventName = string

type TSWConnector_Message struct {
	EventName  TSWConnector_Message_EventName
	Properties map[string]string
}

type TSWConnector interface {
	Start() error
	Stop() error
	IsActive() bool
	Port() int
	Subscribe() (chan TSWConnector_Message, func())
	Send(m TSWConnector_Message) error
}

func TSWConnector_Message_FromString(msg string) TSWConnector_Message {
	parts := strings.Split(msg, ",")
	result := TSWConnector_Message{
		EventName:  "",
		Properties: make(map[string]string),
	}

	if len(parts) == 0 {
		return result
	}

	// first part is the event name
	result.EventName = parts[0]

	// the rest are key=value pairs
	for _, p := range parts[1:] {
		if kv := strings.SplitN(p, "=", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			result.Properties[key] = val
		}
	}

	return result
}

func (msg TSWConnector_Message) ToString() string {
	var sb strings.Builder

	sb.WriteString(msg.EventName)

	keys := make([]string, 0, len(msg.Properties))
	for k := range msg.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		sb.WriteString(",")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", msg.Properties[k]))
	}

	return sb.String()
}
