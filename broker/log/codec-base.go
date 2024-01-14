package log

import (
	"fmt"
	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/proto/log"
	servicecontext "github.com/pydio/cells/v4/common/service/context"
	json "github.com/pydio/cells/v4/common/utils/jsonx"
	"strings"
	"time"
)

// IndexableLog extends default log.LogMessage struct to add index specific methods
type IndexableLog struct {
	Nano            int `bson:"nano"`
	*log.LogMessage `bson:"inline"`
}

// BleveType is interpreted by bleve indexer as the mapping name
func (*IndexableLog) BleveType() string {
	return "log"
}

type baseCodec struct{}

func (b *baseCodec) Marshal(input interface{}) (interface{}, error) {
	var msg *IndexableLog
	switch v := input.(type) {
	case *IndexableLog:
		msg = v
	case *log.Log:
		if ms, e := b.marshalLogMsg(v); e == nil {
			msg = ms
		} else {
			return nil, e
		}
	case *log.LogMessage:
		return &IndexableLog{LogMessage: v}, nil
	default:
		return nil, fmt.Errorf("unrecognized type")
	}
	return msg, nil
}

// marshalLogMsg creates an IndexableLog object and populates the inner LogMessage with known fields of the passed JSON line.
func (b *baseCodec) marshalLogMsg(line *log.Log) (*IndexableLog, error) {

	msg := &IndexableLog{
		LogMessage: &log.LogMessage{},
	}
	zaps := make(map[string]interface{})
	var data map[string]interface{}
	e := json.Unmarshal(line.Message, &data)
	if e != nil {
		return nil, e
	}

	for k, v := range data {
		val, ok := v.(string)
		if !ok && k != common.KeyTransferSize {
			zaps[k] = v
			continue
		}
		switch k {
		case "ts":
			t, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return nil, err
			}
			msg.Ts = int32(t.UTC().Unix())
		case "level":
			msg.Level = val
		case common.KeyMsgId:
			msg.MsgId = val
		case "logger": // name of the service that is currently logging.
			msg.Logger = val
		// N specific info
		case common.KeyNodeUuid:
			msg.NodeUuid = val
		case common.KeyNodePath:
			msg.NodePath = val
		case common.KeyTransferSize:
			if f, o := v.(float64); o {
				msg.TransferSize = int64(f)
			} else if i, o2 := v.(int64); o2 {
				msg.TransferSize = i
			}
		case common.KeyWorkspaceUuid:
			msg.WsUuid = val
		case common.KeyWorkspaceScope:
			msg.WsScope = val
		// User specific info
		case common.KeyUsername:
			msg.UserName = val
		case common.KeyUserUuid:
			msg.UserUuid = val
		case common.KeyGroupPath:
			msg.GroupPath = val
		case common.KeyRoles:
			msg.RoleUuids = strings.Split(val, ",")
		case common.KeyProfile:
			msg.Profile = val
		// Session and remote client info
		case servicecontext.HttpMetaRemoteAddress:
			msg.RemoteAddress = val
		case servicecontext.HttpMetaUserAgent:
			msg.UserAgent = val
		case servicecontext.HttpMetaProtocol:
			msg.HttpProtocol = val
		// Span enable following a given request between the various services
		case common.KeySpanUuid:
			msg.SpanUuid = val
		case common.KeySpanParentUuid:
			msg.SpanParentUuid = val
		case common.KeySpanRootUuid:
			msg.SpanRootUuid = val
		// Group messages for a given high level operation
		case common.KeyOperationUuid:
			msg.OperationUuid = val
		case common.KeyOperationLabel:
			msg.OperationLabel = val
		case common.KeySchedulerJobId:
			msg.SchedulerJobUuid = val
		case common.KeySchedulerTaskId:
			msg.SchedulerTaskUuid = val
		case common.KeySchedulerActionPath:
			msg.SchedulerTaskActionPath = val
		case "msg", "error":
		default:
			zaps[k] = v
		}
	}

	// Concatenate msg and error in the full text msg field.
	text := ""
	if m, ok := data["msg"]; ok {
		if t, o := m.(string); o {
			text = t
		} else {
			fmt.Println("Error while unmarshaling log data, data['msg'] not a string", m)
		}
	}
	if m, ok := data["error"]; ok {
		if stringErr, o := m.(string); o {
			text += " - " + stringErr
		}
	}
	msg.Msg = text
	msg.Nano = int(line.Nano)

	if len(zaps) > 0 {
		data, _ := json.Marshal(zaps)
		msg.JsonZaps = string(data)
	}

	return msg, nil
}
