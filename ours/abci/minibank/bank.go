package minibank

import (
	"emulator/ours/definition"
	"emulator/ours/types"
	bank "emulator/proto/ours/abci/minibank"
	"emulator/utils"
	"emulator/utils/store"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

const initBalance = 1000000

type Application struct {
	db *store.PrefixStore

	KeyRangeTree         *utils.RangeTree
	relatedIKeyRangeTree map[string]*utils.RangeTree
	IsI                  bool
	myChainID            string

	appStatus []byte

	LockedKeys map[string]bool

	preExecutionInputList  map[string][]map[string]uint32
	preExecutionOutputList map[string][]map[string]uint32
}

func NewApplication(dbDir string, chain_id string, isI bool, keyRangeTree *utils.RangeTree, relatedIKeyRangeTree map[string]*utils.RangeTree) *Application {
	app := new(Application)
	app.db = store.NewPrefixStore("abci.minibank", dbDir)
	app.KeyRangeTree, app.relatedIKeyRangeTree = keyRangeTree, relatedIKeyRangeTree

	app.IsI = isI
	app.myChainID = chain_id

	app.appStatus = []byte("minibank")
	app.LockedKeys = map[string]bool{}

	app.preExecutionInputList = map[string][]map[string]uint32{}
	app.preExecutionOutputList = map[string][]map[string]uint32{}
	return app
}
func (app *Application) Stop() {
	app.db.Close()
}

var _ definition.ABCIConn = (*Application)(nil)

func (app *Application) Commit() []byte {
	app.preExecutionInputList = map[string][]map[string]uint32{}
	app.preExecutionOutputList = map[string][]map[string]uint32{}
	
	return app.appStatus
}

func (app *Application) PreExecutionI(blockList []*types.Block) *types.ABCIPreExecutionResponseI {
	fmt.Printf("%v IshardPreexecute\n", time.Now())
	commitStatuses := make([][]byte, len(blockList))
	alldg := utils.NewDependencyGraph()

	for blockIndex, block := range blockList {
		bv := utils.NewBitVector(block.CrossShardTxs.Size())
		dg := utils.NewDependencyGraph()

		for i, tx := range block.CrossShardTxs {
			dgFlag := true
			readSet, writeSet := app.getRWSet(tx)
			for r := range readSet {
				if index, ok := dg.AddRead(r); ok && !bv.GetIndex(index) {
					dgFlag = false
					break
				} else if _, ifPreviousWrite := alldg.AddRead(r); !ok && ifPreviousWrite {
					dgFlag = false
					break
				}
			}
			for w := range writeSet {
				dg.AddWrite(w, i)
			}
			bv.SetIndex(i, dgFlag)
		}
		for w := range dg.Map() {
			alldg.AddWrite(w, blockIndex)
		}
		commitStatuses[blockIndex] = bv.Byte()
	}

	wg := sync.WaitGroup{}
	wg.Add(len(blockList))
	preExecutionBFunc := func(blockIndex int, block *types.Block) {
		defer wg.Done()
		memoryDB := NewMemoryDB(app.db)
		bv := utils.NewBitArrayFromByte(commitStatuses[blockIndex])
		dg := utils.NewDependencyGraph()
		for i, tx := range block.CrossShardTxs {
			if !bv.GetIndex(i) {
				continue
			}
			csdb, ok, err := NewCrossShardDataFromBytes(block.CrossShardDatas[i])
			if err != nil {
				//fmt.Println(i, err)
				bv.SetIndex(i, false)
			} else if !ok {
			
				bv.SetIndex(i, true)
			} else {
				crossShardDataFreshFlag := true
				for key, value := range csdb {
					if app.KeyRangeTree.Search(key) {
						if v, err := memoryDB.Get(key); err != nil || v != value {

							crossShardDataFreshFlag = false
							break
						}
					}
				}
				if !crossShardDataFreshFlag {
					bv.SetIndex(i, false)
				} else {
					db := NewCrossShardDB(memoryDB, csdb)
					err := app.execute([]byte(tx), db)
					if err != nil {
						bv.SetIndex(i, false)
						db.CommitWrite(false)
					} else {
						readSet := db.GetReadSet()
						writeSet := db.GetWriteSet()
						dgFlag := true
						for key := range readSet {
							if index, ok := dg.AddRead(key); ok && !bv.GetIndex(index) {
								dgFlag = false
								break
							}
						}
						for key := range writeSet {
							dg.AddWrite(key, i)
						}
						bv.SetIndex(i, dgFlag)
						db.CommitWrite(dgFlag)
					}
				}
			}
		}
		commitStatuses[blockIndex] = bv.Byte()
	}

	for blockIndex, block := range blockList {
		go preExecutionBFunc(blockIndex, block)
	}
	wg.Wait()
	fmt.Printf("%v Ishard_preexecute end\n", time.Now())
	return &types.ABCIPreExecutionResponseI{
		Code:             types.CodeTypeOK,
		CommitStatusVote: commitStatuses,
	}
}

func (app *Application) PreExecutionB(block *types.Block) *types.ABCIPreExecutionResponseB {
	fmt.Printf("%v Bshard_preexecute\n", time.Now())
	memoryDB := NewMemoryDB(app.db)
	db := NewRangeDB(memoryDB, app.KeyRangeTree)
	datas := make([][]byte, block.CrossShardTxs.Size())
	blockKey := string(block.Hash())

	app.preExecutionInputList[blockKey] = make([]map[string]uint32, 0, block.CrossShardTxs.Size())
	app.preExecutionOutputList[blockKey] = make([]map[string]uint32, 0, block.CrossShardTxs.Size())
	for i, tx := range block.CrossShardBody.CrossShardTxs {
		err := app.execute([]byte(tx), db)
		readSet, writeSet := map[string]uint32{}, map[string]uint32{}
		if err == nil {
			readSet = db.GetReadSet()
			writeSet = db.GetWriteSet()
		}
		app.preExecutionInputList[blockKey] = append(app.preExecutionInputList[blockKey], readSet)
		app.preExecutionOutputList[blockKey] = append(app.preExecutionOutputList[blockKey], writeSet)
		db.CommitWrite(err == nil)
		datas[i] = CrossShardDataBytes(readSet, err == nil)
	}
	fmt.Printf("%v Bshard_preexecuteend\n", time.Now())
	return &types.ABCIPreExecutionResponseB{
		Code:            types.CodeTypeOK,
		CrossShardDatas: datas,
	}
}
func (app *Application) getStatus() {
	log.Println("dbstate:")
	iter, err := app.db.Iterator([]byte(app.KeyRangeTree.StartKey()), []byte(app.KeyRangeTree.EndKey()))
	if err != nil {
		panic(err)
	}
	for iter.Valid() {
		fmt.Println(string(iter.Key()), utils.BytesToUint32(iter.Value()))
		iter.Next()
	}
	iter.Close()
}
func (app *Application) Execution(CrossShardBlocks []*types.Block, commitBitVector []*utils.BitVector, InnerShardBlocks []*types.Block) *types.ABCIExecutionResponse {
	//defer app.getStatus()
	resp := &types.ABCIExecutionResponse{
		CrossShardResps: make(map[string][]*types.ABCIExecutionReceipt),
		InnerShardResps: make(map[string][]*types.ABCIExecutionReceipt),
	}
	memDBlist := make([]*MemoryDB, len(CrossShardBlocks))
	CrossShardTxsCount, CrossShardCommitTxsCount := 0, 0

	crossMutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(CrossShardBlocks))
	executionB := func(block_sequence int, block *types.Block) {
		defer wg.Done()
		var count, commit int
		memoryDB := NewMemoryDB(app.db)
		bv := commitBitVector[block_sequence]
		respSet := make([]*types.ABCIExecutionReceipt, block.CrossShardTxs.Size())
		preExecutionOutput := app.preExecutionOutputList[string(block.Hash())]
		count = bv.Size()
		for i, tx := range block.CrossShardTxs {
			abciResp := new(types.ABCIExecutionReceipt)
			csdb, ok, err := NewCrossShardDataFromBytes(block.CrossShardDatas[i])
			if err != nil {
				abciResp.Code = types.CodeTypeEncodingError
				abciResp.Log = err.Error()
			} else if !ok {
				abciResp.Code = types.CodeTypeAbort
				abciResp.Log = "B PreExecution Abort"
			} else if !bv.GetIndex(i) {
				abciResp.Code = types.CodeTypeAbort
				abciResp.Log = "I PreExecution Abort"
			} else if len(preExecutionOutput) > i {
				for key, value := range preExecutionOutput[i] {
					memoryDB.Set(key, value)
				}
				abciResp.Code = types.CodeTypeOK
				memoryDB.CommitWrite(true)
			} else {
				db := NewCrossShardDB(memoryDB, csdb)
				err := app.execute([]byte(tx), db)
				if err != nil {
					abciResp.Code = types.CodeTypeUnknownError
					abciResp.Log = err.Error()
					panic(err)
				} else {
					abciResp.Code = types.CodeTypeOK
				}
				db.CommitWrite(err == nil)
			}
			respSet[i] = abciResp
			if abciResp.IsOK() {
				commit++
			}
		}
		crossMutex.Lock()
		defer crossMutex.Unlock()
		CrossShardTxsCount += count
		CrossShardCommitTxsCount += commit
		resp.CrossShardResps[block.ChainID] = respSet
		memDBlist[block_sequence] = memoryDB
	}
	for block_sequence, block := range CrossShardBlocks {
		go executionB(block_sequence, block)
	}
	wg.Wait()

	for _, memoryDB := range memDBlist {
		tmp_write_db := NewRangeDB(memoryDB, app.KeyRangeTree)
		if err := tmp_write_db.Write(); err != nil {
			panic(err)
		}
	}
	fmt.Printf("(%v)finish%dshard cross-shard block，total%dtransactions，commit number are%d\n", time.Now(), len(CrossShardBlocks), CrossShardTxsCount, CrossShardCommitTxsCount)

	if app.IsI {
		commitTxsCount := 0
		block := InnerShardBlocks[0]
		mmrDB := NewMemoryDB(app.db)
		db := NewRangeDB(mmrDB, app.KeyRangeTree)
		respSet := make([]*types.ABCIExecutionReceipt, block.BodyTxs.Size())
		for i, tx := range block.BodyTxs {
			abciResp := new(types.ABCIExecutionReceipt)
			err := app.execute([]byte(tx), db)
			if err != nil {
				abciResp.Code = types.CodeTypeAbort
				abciResp.Log = err.Error()
			} else {
				abciResp.Code = types.CodeTypeOK
				commitTxsCount++
			}
			db.CommitWrite(err == nil)
			respSet[i] = abciResp
		}
		resp.InnerShardResps[block.ChainID] = respSet
		fmt.Println("start write", time.Now())
		if err := db.Write(); err != nil {
			panic(err)
		}
		fmt.Printf("(%v)finish Ishard intra-shard process，totl%dtransactions，commit number are%d\n", time.Now(), block.BodyTxs.Size(), commitTxsCount)
	} else {
		commitLock := sync.Mutex{}
		commitWaitGroup := sync.WaitGroup{}
		commitWaitGroup.Add(len(InnerShardBlocks))
		commitTxsCount := 0
		execBlock := func(b *types.Block) {
			defer commitWaitGroup.Done()
			mmrDB := NewMemoryDB(app.db)
			db := NewRangeDB(mmrDB, app.relatedIKeyRangeTree[b.ChainID])
			respSet := make([]*types.ABCIExecutionReceipt, b.BodyTxs.Size())
			cnt := 0
			for i, tx := range b.BodyTxs {
				abciResp := new(types.ABCIExecutionReceipt)
				err := app.execute([]byte(tx), db)
				if err != nil {
					abciResp.Code = types.CodeTypeAbort
					abciResp.Log = err.Error()
				} else {
					abciResp.Code = types.CodeTypeOK
					cnt++
				}
				db.CommitWrite(err == nil)
				respSet[i] = abciResp
			}
			fmt.Println("start write", time.Now())
			if err := db.Write(); err != nil {
				panic(err)
			}
			commitLock.Lock()
			defer commitLock.Unlock()
			commitTxsCount += cnt
			resp.InnerShardResps[b.ChainID] = respSet
			fmt.Printf("(%v)finish%sshard intra-shard process，total%dtransactions，commit transactions are%d\n", time.Now(), b.ChainID, b.BodyTxs.Size(), cnt)
		}
		for _, block := range InnerShardBlocks {
			go execBlock(block)
		}
		commitWaitGroup.Wait()
	}
	return resp
}

func (app *Application) execute(txBytes []byte, db GetSetMemDB) error {
	if len(txBytes) < 4 {
		return errors.New("given tx Length less than 4，Illegal format")
	}
	txType := utils.BytesToUint32(txBytes[:4])
	switch txType {
	case definition.TxTransfer:
		tx, err := NewTransferTxFromBytes(txBytes)
		if err != nil {
			return err
		}
		return app.doTransfer(tx, db)
	case definition.TxInsert:
		tx := new(bank.InsertTx)
		if err := proto.Unmarshal(txBytes[4:], tx); err != nil {
			return err
		}
		return app.doInsert(tx, db)
	default:
		return errors.New("ABCI Tx Undefined type")
	}
}
func (app *Application) doTransfer(tx *bank.TransferTx, db GetSetMemDB) error {
	if err := ValidateTransferTx(tx); err != nil {
		return err
	}
	fromBalance, toBalance := make([]uint32, len(tx.From)), make([]uint32, len(tx.To))
	for i, fromKey := range tx.From {
		if app.LockedKeys[fromKey] {
			return errors.New("The primary key is locked")
		}
		if balance, err := db.Get(fromKey); err != nil {
			return err
			//fromBalance[i] = initBalance - tx.FromMoney[i]
		} else if balance < tx.FromMoney[i] {
			return errors.New("Balance is not Enough")
		} else {
			fromBalance[i] = balance - tx.FromMoney[i]
		}
	}
	for i, toKey := range tx.To {
		if app.LockedKeys[toKey] {
			return errors.New("The primary key is locked")
		}
		if balance, err := db.Get(toKey); err != nil {
			return err
			//toBalance[i] = initBalance + tx.ToMoney[i]
		} else {
			toBalance[i] = balance + tx.ToMoney[i]
		}
	}
	for i, fromKey := range tx.From {
		if app.LockedKeys[fromKey] {
			return errors.New("The primary key is locked")
		}

		if err := db.Set(fromKey, fromBalance[i]); err != nil {
			return err
		}
	}
	for i, toKey := range tx.To {
		if app.LockedKeys[toKey] {
			return errors.New("The primary key is locked")
		}
		if err := db.Set(toKey, toBalance[i]); err != nil {
			return err
		}
	}
	return nil
}

func (app *Application) doInsert(tx *bank.InsertTx, db GetSetMemDB) error {
	if app.LockedKeys[tx.Account] {
		return errors.New("Key Already Locked")
	}
	if ok, err := db.Has(tx.Account); err != nil {
		return err
	} else if ok {
		return errors.New("Key Already Exists")
	}
	if err := db.Set(tx.Account, tx.Money); err != nil {
		return err
	}
	return nil
}

func (app *Application) ValidateTx(txBytes types.Tx, isCrossShard bool) bool {
	if len(txBytes) < 4 {
		fmt.Println("Insufficient transaction length")
		return false
	}
	var rangeTree *utils.RangeTree
	if app.IsI {
		if isCrossShard {
			fmt.Println("Ishard invokes cross-shard Mempool")
			return false
		}
		rangeTree = app.KeyRangeTree
	} else {
		if isCrossShard {
			rangeTree = app.KeyRangeTree
		} else {
			rangeTree = app.relatedIKeyRangeTree[app.myChainID]
		}
	}
	txType := utils.BytesToUint32(txBytes[:4])
	switch txType {
	case definition.TxTransfer:
		tx := new(bank.TransferTx)
		if err := proto.Unmarshal(txBytes[4:], tx); err != nil {
			fmt.Println("Parsing failed")
			return false
		}
		for _, key := range tx.From {
			if !rangeTree.Search(key) {
				fmt.Println(key, rangeTree.Range())
				fmt.Println("RangeTree out of bound")
				return false
			}
		}
		for _, key := range tx.To {
			if !rangeTree.Search(key) {
				fmt.Println(key, rangeTree.Range())
				fmt.Println("RangeTree out of bound")
				return false
			}
		}
		return true
	case definition.TxInsert:
		tx := new(bank.InsertTx)
		if err := proto.Unmarshal(txBytes[4:], tx); err != nil {
			return false
		}
		return rangeTree.Search(tx.Account)
	default:
		fmt.Println("Transaction of unknown type")
		return false
	}
}

func (app *Application) getRWSet(txBz []byte) (map[string]bool, map[string]bool) {
	//start := time.Now()
	tx, err := NewTransferTxFromBytes(txBz)
	//fmt.Println("start Time", time.Since(start))
	if err != nil {
		return map[string]bool{}, map[string]bool{}
	}
	readSet, writeSet := map[string]bool{}, map[string]bool{}
	for _, key := range tx.From {
		writeSet[key] = true
		readSet[key] = true
	}
	for _, key := range tx.To {
		writeSet[key] = true
		readSet[key] = true
	}
	return readSet, writeSet
}
