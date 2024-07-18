package mempool

import (
	"emulator/libs/clist"
	"emulator/uranus/definition"
	"emulator/uranus/types"
	"emulator/utils"
	"fmt"
	"sync"
)

type Mempool struct {
	txs    *clist.CList
	txsMap sync.Map

	abci                definition.ABCIConn
	isCrossShardMempool bool
}

func NewMempool(isCrossShardMempool bool, abci definition.ABCIConn) *Mempool {
	return &Mempool{
		txs:    clist.New(),
		txsMap: sync.Map{},

		abci:                abci,
		isCrossShardMempool: isCrossShardMempool,
	}
}

var _ definition.MempoolConn = (*Mempool)(nil)

func (mpl *Mempool) Receive(chID byte, message []byte, messageType uint32) error {
	tx := types.Tx(message)
	return mpl.AddTx(&tx)
}

func (mpl *Mempool) AddTx(tx *types.Tx) error {
	if !mpl.abci.ValidateTx(*tx, mpl.isCrossShardMempool) {
		return fmt.Errorf("The transaction did not pass ValidateTx")
	}
	e := mpl.txs.PushBack(tx)
	mpl.txsMap.Store(tx.Key(), e)
	return nil
}

func (mpl *Mempool) RemoveTx(tx *types.Tx) error {
	if e, ok := mpl.txsMap.Load(tx.Key()); ok {
		if u, ok := e.(*clist.CElement); ok {
			mpl.txs.Remove(u)
			mpl.txsMap.Delete(tx.Key())
		}
	}
	return nil
}

func (mpl *Mempool) ReapTx(maxTxs int) (types.Txs, int, error) {
	var (
		txs = make(types.Txs, 0, maxTxs)
	)
	for e := mpl.txs.Front(); e != nil; e = e.Next() {
		memTx := *e.Value.(*types.Tx)
		txs = append(txs, memTx)

		if len(txs) >= maxTxs {
			break
		}
	}
	return txs, len(txs), nil
}

func (mpl *Mempool) Update(txs types.Txs, commitStatus []byte) error {
	if len(txs) == 0 {
		return nil
	}
	if mpl.isCrossShardMempool {
		bv := utils.NewBitArrayFromByte(commitStatus)
		for i, tx := range txs {
			if bv.GetIndex(i) {
				mpl.RemoveTx(&tx)
			}
		}
		return nil
	} else {
		for _, tx := range txs {
			mpl.RemoveTx(&tx)
		}
		return nil
	}

}
