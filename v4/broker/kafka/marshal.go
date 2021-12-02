package kafka

import (
	"encoding/json"

	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/codec/bytes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// Marshaler is a simple encoding interface used for the broker/transport
// where headers are not supported by the underlying implementation.
type Marshaler struct{}

// Marshal returns the JSON encoding of v
func (Marshaler) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case *bytes.Frame:
		return m.Data, nil
	case proto.Message:
		return proto.Marshal(m)
	case protoiface.MessageV1:
		m2 := protoimpl.X.ProtoMessageV2Of(m)
		return proto.Marshal(m2)
	}
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
func (Marshaler) Unmarshal(d []byte, v interface{}) error {
	switch m := v.(type) {
	case proto.Message:
		return protojson.Unmarshal(d, m)
	case protoiface.MessageV1:
		m2 := protoimpl.X.ProtoMessageV2Of(m)
		return protojson.Unmarshal(d, m2)
	case *broker.Message:
		if m.Header == nil {
			m.Header = make(map[string]string)
		}
		m.Header["Content-Type"] = "application/json"
		m.Body = d
		return nil
	}
	return json.Unmarshal(d, v)
}

func (Marshaler) String() string {
	return "json-marshaler"
}
