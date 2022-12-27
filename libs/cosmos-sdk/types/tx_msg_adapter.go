package types

import (
	"github.com/gogo/protobuf/proto"
	"github.com/evoblockchain/evochain/libs/cosmos-sdk/codec"
)

const (
	IBCROUTER = "ibc"
)

type MsgProtoAdapter interface {
	Msg
	codec.ProtoMarshaler
}
type MsgAdapter interface {
	Msg
	proto.Message
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`.
func MsgTypeURL(msg MsgProtoAdapter) string {
	return "/" + proto.MessageName(msg)
}
