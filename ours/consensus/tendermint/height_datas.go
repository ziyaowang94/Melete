package tendermint

import (
	"bytes"
	"emulator/ours/consensus/tendermint/constypes"
	"emulator/ours/definition"
	"emulator/ours/shardinfo"
	"emulator/ours/types"
	"emulator/utils"
	"emulator/utils/p2p"
	crypto "emulator/utils/signer"
	"fmt"
	"log"
)

var (
	DoNothing = fmt.Errorf("")
)

type HeightDataPackage struct {
	Height    int64
	Round     int
	MyChainID string
	PassStep  int8
	state     *ConsensusState

	ShardInfo       *shardinfo.ShardInfo
	ShardVotes      map[string]int
	AllValidatorSet map[string]*crypto.Verifier

	Proposal *constypes.Proposal

	Prevotes          []*constypes.Prevote
	PrevotesBitVector *utils.BitVector
	PrevotesTotal     int

	Precommits          []*constypes.Precommit
	PrecommitsBitVector *utils.BitVector
	PrecommitsTotal     int

	CrossShardProposals map[string]*constypes.CrossShardProposal

	CrossShardAccepts             map[string]*constypes.CSAShardSet
	CrossShardAcceptFinishCounter int

	CrossShardCommits map[string]*constypes.CrossShardCommit

	CrossShardPartSets map[string]*types.PartSet
	CrossShardBlocks   map[string]*types.Block

	randomHeaderBuffer map[string][]*types.Part
	p2pConn            *p2p.Sender
}

func (h *HeightDataPackage) IsIShard() bool { return h.ShardInfo.IsIShard() }
func (h *HeightDataPackage) IsBShard() bool { return h.ShardInfo.IsBShard() }


func NewHeightData(mychainid string, height int64, si *shardinfo.ShardInfo, sender *p2p.Sender) *HeightDataPackage {
	h := new(HeightDataPackage)
	h.Height, h.Round = height, 0
	h.PassStep = RoundStepNewHeight
	h.MyChainID = mychainid
	h.p2pConn = sender
	h.SetShardInfo(si)
	h.clearAll()
	return h
}
func (h *HeightDataPackage) NextHeight() {
	h.Height++
	h.Round = 0
	h.PassStep = RoundStepNewHeight
	h.clearAll()
}
func (h *HeightDataPackage) setKeys(pubkeys []string, chain_id string) error {
	verifier, err := crypto.NewVerifier(pubkeys)
	if err != nil {
		return err
	} else if verifier == nil {
		return fmt.Errorf("HeightData initialization: empty Validator error")
	}
	h.AllValidatorSet[chain_id] = verifier
	//h.AllValidatorSetSize[chain_id] = len(pubkeys)
	return nil
}

func (h *HeightDataPackage) SetShardInfo(s *shardinfo.ShardInfo) error {
	h.ShardInfo = s
	h.AllValidatorSet = make(map[string]*crypto.Verifier)
	h.ShardVotes = make(map[string]int)
	//h.AllValidatorSetSize = make(map[string]int)
	for chain_id, peers := range s.PeerList {
		totalVotes := 0
		peerPubKeys := make([]string, len(peers))
		for i, peer := range peers {
			peerPubKeys[i] = peer.Pubkey
			totalVotes += int(peer.Vote)
		}
		if err := h.setKeys(peerPubKeys, chain_id); err != nil {
			return err
		}
		h.ShardVotes[chain_id] = totalVotes
	}
	return nil
}



func (h *HeightDataPackage) FinishProposal() bool {
	if h.PassStep >= RoundStepPropose {
		return true
	}
	if h.Proposal == nil {
		return false
	}
	myid := h.MyChainID
	if b, ok := h.CrossShardBlocks[myid]; !ok || b == nil {
		return false
	}

	h.PassStep = RoundStepPropose
	return true
}
func (h *HeightDataPackage) FinishPrevote() bool {
	if h.PassStep >= RoundStepPrevote {
		return true
	}
	if !h.FinishProposal() {
		return false
	}
	if !(h.PrevotesTotal > h.ShardVotes[h.MyChainID]*2/3) {
		return false
	}
	h.PassStep = RoundStepPrevote
	return true
}
func (h *HeightDataPackage) FinishPrecommit() bool {
	if h.PassStep >= RoundStepPrecommit {
		return true
	}
	if !h.FinishPrevote() {
		return false
	}
	if !(h.PrecommitsTotal > h.ShardVotes[h.MyChainID]*2/3) {
		return false
	}
	h.PassStep = RoundStepPrecommit
	return true
}

func (h *HeightDataPackage) FinishCrossShardAccept() bool {
	if h.PassStep >= RoundStepCrossShardAccept {
		return true
	}
	if !h.FinishPrecommit() {
		return false
	}
	ProposalSize := len(h.ShardInfo.RelatedShards)
	if len(h.CrossShardProposals) < ProposalSize+1 {
		return false
	}
	if len(h.CrossShardBlocks) < ProposalSize+1 {
		return false
	}
	for key := range h.ShardInfo.RelatedShards {
		if _, ok := h.CrossShardProposals[key]; !ok {
			return false
		}
		if _, ok := h.CrossShardBlocks[key]; !ok {
			return false
		}
	}
	h.PassStep = RoundStepCrossShardAccept
	return true
}

func (h *HeightDataPackage) FinishCrossShardProposal() bool {
	if h.PassStep >= RoundStepCrossShardPropose {
		return true
	}
	if !h.FinishPrecommit() {
		return false
	}
	//fmt.Println("CSA Shards", h.CrossShardAcceptFinishCounter, len(h.ShardInfo.RelatedShards))
	if h.CrossShardAcceptFinishCounter < len(h.ShardInfo.RelatedShards) {
		return false
	}
	h.PassStep = RoundStepCrossShardPropose
	return true
}
func (h *HeightDataPackage) FinishCrossShardCommit_I() bool {
	if h.PassStep >= RoundStepCrossShardCommit {
		return true
	}
	if !h.FinishCrossShardAccept() {
		return false
	}
	if len(h.CrossShardCommits) < len(h.ShardInfo.RelatedShards) {
		return false
	}
	for key := range h.ShardInfo.RelatedShards {
		if _, ok := h.CrossShardCommits[key]; !ok {
			return false
		}
	}
	h.PassStep = RoundStepCrossShardCommit
	return true
}
func (h *HeightDataPackage) FinishCrossShardCommit_B() bool {
	if h.PassStep >= RoundStepCrossShardCommit {
		return true
	}
	if !h.FinishCrossShardProposal() {
		return false
	}

	if len(h.CrossShardCommits) < len(h.ShardInfo.Related2LevelShards)+1 {
		return false
	}

	if len(h.CrossShardProposals) < len(h.ShardInfo.RelatedShards)+len(h.ShardInfo.Related2LevelShards)+1 {
		return false
	}
	if len(h.CrossShardBlocks) < len(h.ShardInfo.RelatedShards)+len(h.ShardInfo.Related2LevelShards)+1 {
		return false
	}
	for key := range h.ShardInfo.RelatedShards {
		if _, ok := h.CrossShardBlocks[key]; !ok {
			return false
		}
		if _, ok := h.CrossShardProposals[key]; !ok {
			return false
		}
	}
	for key := range h.ShardInfo.Related2LevelShards {
		if _, ok := h.CrossShardBlocks[key]; !ok {
			return false
		}
		if _, ok := h.CrossShardProposals[key]; !ok {
			return false
		}
		if _, ok := h.CrossShardCommits[key]; !ok {
			return false
		}
	}
	return true
}



func (h *HeightDataPackage) clearProposal() {
	h.Proposal = nil
	delete(h.CrossShardPartSets, h.MyChainID)
	delete(h.CrossShardBlocks, h.MyChainID)
}
func (h *HeightDataPackage) clearPrevote() {
	size := h.AllValidatorSet[h.MyChainID].Size()
	h.Prevotes = make([]*constypes.Prevote, size)
	h.PrevotesBitVector = utils.NewBitVector(size)
	h.PrevotesTotal = 0
}
func (h *HeightDataPackage) clearPrecommit() {
	size := h.AllValidatorSet[h.MyChainID].Size()
	h.Precommits = make([]*constypes.Precommit, size)
	h.PrecommitsBitVector = utils.NewBitVector(size)
	h.PrecommitsTotal = 0
}
func (h *HeightDataPackage) clearCrossShardData() {
	h.CrossShardProposals = make(map[string]*constypes.CrossShardProposal)

	h.CrossShardAccepts = make(map[string]*constypes.CSAShardSet)
	for cid := range h.ShardInfo.RelatedShards {
		h.CrossShardAccepts[cid] = constypes.NewCSAShardSet(h.Height, cid, h.ShardInfo.PeerList[cid], h.AllValidatorSet[cid])
	}
	h.CrossShardAcceptFinishCounter = 0

	h.CrossShardBlocks = make(map[string]*types.Block)
	h.CrossShardCommits = make(map[string]*constypes.CrossShardCommit)
	h.CrossShardPartSets = make(map[string]*types.PartSet)

	h.randomHeaderBuffer = make(map[string][]*types.Part)
}
func (h *HeightDataPackage) clearAll() {
	h.clearProposal()
	h.clearPrevote()
	h.clearPrecommit()
	h.clearCrossShardData()
}



func (h *HeightDataPackage) AddProposal(p *constypes.Proposal, index int) error {
	if h.Proposal != nil {
		return DoNothing
	}
	if h.Height != p.Header.Height || h.Round != p.Header.Round {
		return DoNothing
	}
	if p.ProposerIndex != index {
		return fmt.Errorf(fmt.Sprintf("Proposer error: Should have been %d, received %d", index, p.ProposerIndex))
	}
	if p.Header.ChainID != h.MyChainID {
		return fmt.Errorf(fmt.Sprintf("Proposal validation failed: I am  %s shard and do not accept the Proposal for %s shard", h.MyChainID, p.Header.ChainID))
	}
	if !h.AllValidatorSet[h.MyChainID].Verify(p.Signature, p.SignBytes(), index) {
		return fmt.Errorf(fmt.Sprintf("Proposal signature verification failed"))
	}
	h.Proposal = p
	h.CrossShardPartSets[h.MyChainID] = types.NewPartSet(p.Header, p.BlockHeaderHash)
	if parts, ok := h.randomHeaderBuffer[p.Header.ChainID]; ok {
		for _, part := range parts {
			h.addInnerPart(part)
		}
	}
	h.reCalculatePrevotes()
	h.reCalculatePrecommits()
	return nil
}

func (h *HeightDataPackage) addInnerPart(p *types.Part) error {
	if h.Height != p.Height || h.Round != p.Round {
		return DoNothing
	}
	if part_set, ok := h.CrossShardPartSets[p.ChainID]; !ok {
		h.randomHeaderBuffer[p.ChainID] = append(h.randomHeaderBuffer[p.ChainID], p)
		return DoNothing
	} else if err := part_set.AddPart(p); err != nil && err != types.DuplicatedPartPassError {
		return err
	} else if err == types.DuplicatedPartPassError {
		return DoNothing
	} else {
		if part_set.IsComplete() {
			block, err := part_set.GenBlock()
			if err != nil {
				return err
			}
			log.Printf("(height=%d)分片%s区块收集完毕\n", h.Height, p.ChainID)
			h.CrossShardBlocks[p.ChainID] = block
		}
	}
	return nil
}

func (h *HeightDataPackage) AddPart(p *types.Part) error {
	return h.addInnerPart(p)
}

func (h *HeightDataPackage) AddPrevote(p *constypes.Prevote) error {
	if h.PrevotesBitVector.GetIndex(p.ValidatorIndex) {
		return DoNothing
	}
	if h.Height != p.Height || h.Round != p.Round {
		return DoNothing
	}
	if maxSize := h.AllValidatorSet[h.MyChainID].Size(); p.ValidatorIndex < 0 || p.ValidatorIndex >= maxSize {
		return fmt.Errorf(fmt.Sprintf("Proposer error: Should be at [0,%d], received %d", maxSize-1, p.ValidatorIndex))
	}

	if !h.AllValidatorSet[h.MyChainID].Verify(p.Signature, p.SignBytes(), p.ValidatorIndex) {
		return fmt.Errorf(fmt.Sprint("Prevote signature verification failed"))
	}
	index := p.ValidatorIndex
	h.Prevotes[index] = p
	if h.Proposal != nil && bytes.Equal(h.Proposal.BlockHeaderHash, p.BlockHeaderHash) {
		h.PrevotesBitVector.SetIndex(index, true)
		h.PrevotesTotal += int(h.ShardInfo.PeerList[h.MyChainID][index].Vote)
	}
	return nil
}

func (h *HeightDataPackage) AddPrecommit(p *constypes.Precommit) error {
	if h.PrecommitsBitVector.GetIndex(p.ValidatorIndex) {
		return DoNothing
	}
	if h.Height != p.Height || h.Round != p.Round {
		return DoNothing
	}
	if maxSize := h.AllValidatorSet[h.MyChainID].Size(); p.ValidatorIndex < 0 || p.ValidatorIndex >= maxSize {
		return fmt.Errorf(fmt.Sprintf("Proposer error: Should have been at [0,%d], received %d", maxSize-1, p.ValidatorIndex))
	}
	if !h.AllValidatorSet[h.MyChainID].Verify(p.Signature, p.SignBytes(), p.ValidatorIndex) {
		return fmt.Errorf(fmt.Sprintf("Precommit signature validation failed"))
	}
	index := p.ValidatorIndex
	h.Precommits[index] = p
	if h.Proposal != nil && bytes.Equal(h.Proposal.BlockHeaderHash, p.BlockHeaderHash) {
		h.PrecommitsBitVector.SetIndex(index, true)
		h.PrecommitsTotal += int(h.ShardInfo.PeerList[h.MyChainID][index].Vote)
	}
	return nil
}

func (h *HeightDataPackage) AddCrossShardProposal(csp *constypes.CrossShardProposal) error {
	chain_id := csp.AggregatedSignatures.Header.ChainID
	if csp.AggregatedSignatures.Header.Height != h.Height {
		return DoNothing
	}
	if _, ok := h.CrossShardProposals[chain_id]; ok {
		return DoNothing
	}
	if !csp.AggregatedSignatures.VerifySignatures(h.AllValidatorSet[chain_id], csp.AggregatedSignatures.ValidatorBitVector.Byte()) {
		return fmt.Errorf(fmt.Sprintf("CrossShardProposal signature verification failed:%s", chain_id))
	}
	var voteTotal int32 = 0
	var bv = csp.AggregatedSignatures.ValidatorBitVector
	for i := 0; i < bv.Size(); i++ {
		if bv.GetIndex(i) {
			voteTotal += h.ShardInfo.PeerList[chain_id][i].Vote
		}
	}
	if !(voteTotal > int32(h.ShardVotes[chain_id]*2/3)) {
		return fmt.Errorf(fmt.Sprintf("CrossShardProposal received less than 2/3 votes:%s", chain_id))
	}

	h.CrossShardProposals[chain_id] = csp
	h.CrossShardPartSets[chain_id] = types.NewPartSet(csp.AggregatedSignatures.Header, csp.AggregatedSignatures.BlockHeaderHash)
	if parts, ok := h.randomHeaderBuffer[chain_id]; ok {
		for _, part := range parts {
			h.AddPart(part)
		}
	}
	return nil
}

func (h *HeightDataPackage) AddCrossShardAccept(csa *constypes.CrossShardAccept) error {
	if csa.Height != h.Height {
		return DoNothing
	}
	if csa.UpperChainID != h.MyChainID {
		return fmt.Errorf(fmt.Sprintf("This is a vote for %s's CrossShardAccept. I'm %s", csa.UpperChainID, h.MyChainID))
	}
	if !h.ShardInfo.RelatedShards[csa.AcceptChainID] {
		return fmt.Errorf(fmt.Sprintf("%s is not my bridged i-shard (%s)", csa.AcceptChainID, h.MyChainID))
	}
	acid, ok := h.CrossShardAccepts[csa.AcceptChainID]
	if !ok {
		return fmt.Errorf(fmt.Sprintf("%s is not my lower shard", csa.AcceptChainID))
	}
	startFlag := acid.HasMaj23()
	if err := acid.AddCSA(csa); err != nil {
		return err
	}
	endFlag := acid.HasMaj23()
	if !startFlag && endFlag {
		h.CrossShardAcceptFinishCounter++
		log.Printf("(height=%d) I received a vote from a committee of CrossShardAccept for the %s shard", h.Height, csa.AcceptChainID)
		fmt.Println(csa.CommitStatus)
	}
	return nil
}

func (h *HeightDataPackage) AddCrossShardCommit(csc *constypes.CrossShardCommit) error {
	if csc.Height != h.Height {
		return DoNothing
	}
	if _, ok := h.CrossShardCommits[csc.UpperChainID]; ok {
		return DoNothing
	}
	if !h.ShardInfo.RelatedShards[csc.UpperChainID] && !h.ShardInfo.Related2LevelShards[csc.UpperChainID] {
		return fmt.Errorf(fmt.Sprintf("This is a vote for %s's CrossShardCommit, not my neighboring shards", csc.UpperChainID))
	}
	if len(csc.ChainList) != len(h.ShardInfo.PeerRelatedMap[csc.UpperChainID]) {
		return fmt.Errorf(fmt.Sprintf("The number of shards in CrossShardCommit does not match the  graph"))
	}
	for i, chain_id := range h.ShardInfo.PeerRelatedMap[csc.UpperChainID] {
		if chain_id != csc.ChainList[i] {
			return fmt.Errorf(fmt.Sprintf("CrossShardCommit's ChainID does not match"))
		}
		if err := csc.ValidateVote(h.ShardInfo.PeerList[chain_id], h.AllValidatorSet[chain_id], i, int32(h.ShardVotes[chain_id])); err != nil {
			return err
		}
	}
	h.CrossShardCommits[csc.UpperChainID] = csc

	if h.IsIShard() {
		relatedMap := make(map[string]struct{}, len(h.ShardInfo.RelatedShards)-1)
		for key := range h.ShardInfo.RelatedShards {
			if key != csc.UpperChainID {
				relatedMap[key] = struct{}{}
			}
		}
		go h.SendToRelatedShardsExcept(csc.ProtoBytes(), definition.TendermintCrossShardCommit, csc.UpperChainID, relatedMap)
	}

	return nil
}

// ================================================================================
func (h *HeightDataPackage) reCalculatePrevotes() {
	if h.Proposal == nil {
		return
	}
	blockHeaderHash := h.Proposal.BlockHeaderHash
	for i, prevote := range h.Prevotes {
		if prevote == nil {
			continue
		}
		if bytes.Equal(blockHeaderHash, prevote.BlockHeaderHash) {
			h.PrevotesBitVector.SetIndex(i, true)
			h.PrevotesTotal += int(h.ShardInfo.PeerList[h.MyChainID][i].Vote)
		}
	}
}

func (h *HeightDataPackage) reCalculatePrecommits() {
	if h.Proposal == nil {
		return
	}
	blockHeaderHash := h.Proposal.BlockHeaderHash
	for i, precommit := range h.Precommits {
		if precommit == nil {
			continue
		}
		if bytes.Equal(blockHeaderHash, precommit.BlockHeaderHash) {
			h.PrecommitsBitVector.SetIndex(i, true)
			h.PrecommitsTotal += int(h.ShardInfo.PeerList[h.MyChainID][i].Vote)
		}
	}
}

// ===================================================================================
func (h *HeightDataPackage) GenerateAggregatePrecommitSignature() *constypes.PrecommitAggregated {
	aggSigs := []string{}
	for i := 0; i < h.PrecommitsBitVector.Size(); i++ {
		if h.PrecommitsBitVector.GetIndex(i) {
			aggSigs = append(aggSigs, h.Precommits[i].Signature)
		}
	}
	aggSig, err := crypto.AggregateSignatures(aggSigs)
	if err != nil {
		panic(err)
	}
	return constypes.NewPrecommitAggregated(h.PrecommitsBitVector.Byte(),
		h.Proposal.Header,
		h.Proposal.BlockHeaderHash,
		aggSig)
}

func (h *HeightDataPackage) GenerateCrossShardCommit() *constypes.CrossShardCommit {
	var (
		size                    = len(h.ShardInfo.RelatedShards)
		chainList               = make([]string, 0, size)
		aggregatedSignatureList = make([]string, 0, size)
		validatorBitVectorList  = make([][]byte, 0, size)
		commitStatusList        = make([][]byte, 0, size)
	)
	for _, key := range h.ShardInfo.PeerRelatedMap[h.MyChainID] {
		// 1
		chainList = append(chainList, key)
		cmtstatus, votes := h.CrossShardAccepts[key].GetMaj23()
		bv := utils.NewBitVector(len(h.ShardInfo.PeerList[key]))
		aggSigList := make([]string, 0, len(votes))
		for _, vote := range votes {
			aggSigList = append(aggSigList, vote.Signature)
			bv.SetIndex(vote.ValidatorIndex, true)
		}
		aggSig, err := crypto.AggregateSignatures(aggSigList)
		if err != nil {
			panic(err)
		}
		// 2
		aggregatedSignatureList = append(aggregatedSignatureList, aggSig)
		// 3
		validatorBitVectorList = append(validatorBitVectorList, bv.Byte())
		// 4
		commitStatusList = append(commitStatusList, cmtstatus)
	}

	return constypes.NewCrossShardCommit(h.MyChainID, h.Height, h.Proposal.BlockHeaderHash, chainList, aggregatedSignatureList, validatorBitVectorList, commitStatusList)
}

func (h *HeightDataPackage) SendToRelatedShardsExcept(bz []byte, messageType uint32, exception string, relatedShards map[string]struct{}) {
	for shard := range relatedShards {
		h.p2pConn.SendToShard(shard, p2p.ChannelIDConsensusState, bz, messageType)
	}
}
