package main

import (
	"emulator/logger/blocklogger"
	"emulator/utils"
	"emulator/utils/store"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"emulator/melete/abci/minibank"
	"emulator/melete/types"
)

func getLastFolderName(path string) string {
	return filepath.Base(path)
}
func main() {
	// ./melete-latency ./mytestnet/127.0.0.1/node1  b1
	rootPath := os.Args[1]
	storePath := path.Join(rootPath, "database")
	nodeName := getLastFolderName(rootPath)
	blockloggerDir := path.Join(rootPath, fmt.Sprintf("%s-blocklogger-brief.txt", nodeName))

	reader := blocklogger.NewReader(blockloggerDir)
	blockRangeA, blockRangeB, err := reader.NoneZeroPeriods()
	if err != nil {
		panic(err)
	}

	db := store.NewPrefixStore("consensus", storePath)
	defer db.Close()

	chain_id := os.Args[2]

	innerShardTxCount := 0
	crossShardTxCount := 0
	innerShardLatencyCount := time.Duration(0)
	crossShardLatencyCount := time.Duration(0)

	for i := 0; i < len(blockRangeA); i++ {
		start, end := blockRangeA[i], blockRangeB[i]
		for j := start; j <= end; j++ {
			var block, blockNext *types.Block
			if bz, err := db.GetBlockByHeight(int64(j), chain_id); err != nil {
				continue
			} else if block = types.NewBlockFromBytes(bz); block == nil {
				continue
			}
			if bz, err := db.GetBlockByHeight(int64(j+1), chain_id); err != nil || bz == nil {
				continue
			} else if blockNext = types.NewBlockFromBytes(bz); blockNext == nil {
				continue
			}
			commitTime := blockNext.Time

			for _, txBytes := range block.BodyTxs {
				tx, err := minibank.NewTransferTxFromBytes(txBytes)
				if err != nil {
					continue
				}
				innerShardLatencyCount += commitTime.Sub(utils.ThirdPartyUnmarshalTime(tx.Time))
				innerShardTxCount++
			}
			bv := utils.NewBitArrayFromByte(blockNext.MyCommitStatus)
			for i, txBytes := range block.CrossShardTxs {
				if !bv.GetIndex(i) {
					continue
				}
				tx, err := minibank.NewTransferTxFromBytes(txBytes)
				if err != nil {
					continue
				}
				crossShardLatencyCount += commitTime.Sub(utils.ThirdPartyUnmarshalTime(tx.Time))
				crossShardTxCount++
			}
		}
	}

	if innerShardTxCount > 0 {
		fmt.Printf("intra-shard: %d  avg latency：%.3f sec\n", innerShardTxCount, innerShardLatencyCount.Seconds()/float64(innerShardTxCount))
	}
	if crossShardTxCount > 0 {
		fmt.Printf("cross-shard: %d  avg latency: %.3f sec\n", crossShardTxCount, crossShardLatencyCount.Seconds()/float64(crossShardTxCount))
	}
	if innerShardTxCount+crossShardTxCount > 0 {
		fmt.Printf("all trasactions: %d    avg latency %.3f sec\n",
			crossShardTxCount+innerShardTxCount,
			(crossShardLatencyCount.Seconds()+innerShardLatencyCount.Seconds())/float64(innerShardTxCount+crossShardTxCount),
		)
	}
}
