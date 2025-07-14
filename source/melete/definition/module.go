package definition

import (
	"emulator/melete/types"
	"emulator/utils"
	"emulator/utils/p2p"
)

type MempoolConn interface {
	AddTx(*types.Tx) error
	ReapTx(maxBytes int) (types.Txs, int, error)
	Update(txs types.Txs, commitStatus []byte) error
	RemoveTx(tx *types.Tx) error

	p2p.Reactor
}

type ABCIConn interface {
	ValidateTx(tx types.Tx, isCrossShard bool) bool

	PreExecutionC(*types.Block) *types.ABCIPreExecutionResponseC
	PreExecutionI([]*types.Block) *types.ABCIPreExecutionResponseI

	Execution([]*types.Block, []*utils.BitVector, []*types.Block) *types.ABCIExecutionResponse

	Commit() []byte

	Stop()
}

type ConsensusConn interface {
	p2p.Reactor
	Start()
	Stop()
}
