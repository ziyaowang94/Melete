package definition

import "google.golang.org/protobuf/reflect/protoreflect"

type ProtoMessage interface {
	protoreflect.ProtoMessage
}

const (
	MessageTypeNil = iota
	
	Part
	Block
	BlockHeader

	TendermintProposal
	TendermintPrevote
	TendermintPrecommit
	TendermintCrossShardProposal
	TendermintCrossShardAccept
	TendermintCrossShardCommit

	TxTransfer
	TxInsert
)
