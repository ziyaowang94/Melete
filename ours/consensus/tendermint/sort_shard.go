package tendermint

import "sort"

func ShardIDLess(a, b string) bool {
	return a < b
}

func (cs *ConsensusState) sortShards(shards []string) {
	sort.Slice(shards, func(i, j int) bool {
		return ShardIDLess(shards[i], shards[j])
	})
}
