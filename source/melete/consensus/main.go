package tendermint

import (
	"bytes"
	"emulator/crypto/merkle"
	"emulator/melete/consensus/constypes"
	"emulator/melete/definition"
	inter "emulator/melete/definition"
	"emulator/melete/shardinfo"
	"emulator/melete/types"
	"emulator/utils"
	"emulator/utils/p2p"
	"emulator/utils/store"
	"fmt"
	"log"
	"sync"
	"time"

	"emulator/logger/blocklogger"
	sig "emulator/utils/signer"
)

type ConsensusState struct {
	Height int64 `json:"height"`
	Round  int   `json:"round"`
	Step   int8  `json:"step"`

	mempool             inter.MempoolConn
	cross_shard_mempool inter.MempoolConn
	abci                inter.ABCIConn
	p2p                 *p2p.Sender
	store               *store.PrefixStore

	signer      *sig.Signer
	signerIndex int

	heightDatas         *HeightDataPackage
	relayMessageChannel map[int64]map[int][]interface{}

	MinBlockInterval time.Duration
	LastBlockTime    time.Time
	MaxPartSize      int
	maxBlockTxNum    int

	LastBlockHash           []byte
	LastChainedHash         [][]byte
	LastRelatedCommitStatus [][]byte
	LastCommitStatus        []byte

	LastReceiptRoot []byte
	LastStateRoot   []byte

	BlockHash           []byte
	ChainedHash         [][]byte
	RelatedCommitStatus [][]byte
	CommitStatus        []byte

	logger blocklogger.BlockWriter

	stateMtx sync.Mutex
}

func NewConsensusState(chain_id string, si *shardinfo.ShardInfo,
	signer *sig.Signer, signerIndex int,
	mempool inter.MempoolConn, cross_shard_mempool inter.MempoolConn,
	abci inter.ABCIConn, sender *p2p.Sender,
	storeDir string, minBlockInterval time.Duration, maxPartSize int, maxBlockTxNum int,
	logger blocklogger.BlockWriter) *ConsensusState {
	cs := &ConsensusState{
		Height: 0,
		Round:  0,
		Step:   RoundStepNewHeight,

		mempool:             mempool,
		cross_shard_mempool: cross_shard_mempool,
		abci:                abci,
		p2p:                 sender,
		store:               store.NewPrefixStore("consensus", storeDir),

		signer:      signer,
		signerIndex: signerIndex,

		heightDatas:         NewHeightData(chain_id, 0, si, sender),
		relayMessageChannel: make(map[int64]map[int][]interface{}),

		MinBlockInterval: minBlockInterval,
		LastBlockTime:    time.Now(),
		MaxPartSize:      maxPartSize,
		maxBlockTxNum:    maxBlockTxNum,

		LastBlockHash:           merkle.HashFromByteSlices(nil),
		LastChainedHash:         [][]byte{},
		LastRelatedCommitStatus: [][]byte{},
		LastCommitStatus:        []byte{},

		LastReceiptRoot: merkle.HashFromByteSlices(nil),
		LastStateRoot:   []byte{},

		BlockHash:           []byte{},
		ChainedHash:         [][]byte{},
		RelatedCommitStatus: [][]byte{},
		CommitStatus:        []byte{},

		logger: logger,

		stateMtx: sync.Mutex{},
	}
	cs.heightDatas.state = cs
	return cs
}

func (cs *ConsensusState) Start() {
	//fmt.Println(cs.Height)
	cs.stateMtx.Lock()
	defer cs.stateMtx.Unlock()
	if cs.Height == 0 {
		cs.enterNewHeight()
	} else {
		cs.handleStateTransition()
	}
}
func (cs *ConsensusState) Stop() {
	cs.store.Close()
}
func (cs *ConsensusState) WriteLogger(msg string, is_start, is_end bool) {
	cs.logger.Write(blocklogger.NewConsensusEvent(cs.Height, int32(cs.Round), RoundStepString(cs.Step), is_start, is_end, msg))
}

func (cs *ConsensusState) IsIShard() bool { return cs.heightDatas.IsIShard() }
func (cs *ConsensusState) IsBShard() bool { return cs.heightDatas.IsBShard() }

func (cs *ConsensusState) addRelayMessageIfOverTime(height int64, round int, msg interface{}) bool {
	if cs.Height < height || cs.Height == height && cs.Round < round {
		if u, ok := cs.relayMessageChannel[height]; ok {
			u[round] = append(u[round], msg)
			return true
		} else {
			cs.relayMessageChannel[height] = map[int][]interface{}{round: {msg}}
		}
	}
	return false
}

func (cs *ConsensusState) Next() {
	if cs.Step == RoundStepCrossShardCommit {
		delete(cs.relayMessageChannel, cs.Height)
		cs.Round = 0
		cs.Height++
		cs.Step = RoundStepNewHeight
	} else {
		cs.Step++
	}
}

func (cs *ConsensusState) Receive(channel_id byte, bz []byte, messageType uint32) error {
	switch messageType {
	case definition.Part:
		part := types.NewPartFromBytes(bz)
		if part == nil {
			return fmt.Errorf("Part Unmarshal Error")
		}
		if err := part.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(part)
		return err
	case definition.TendermintProposal:
		proposal := constypes.NewProposalFromBytes(bz)
		if proposal == nil {
			return fmt.Errorf("Proposal Unmarshal Error")
		}

		if err := proposal.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s) Proposal received (height=%d,chain=%s,index=%d)\n", time.Now(), cs.Height, RoundStepString(cs.Step), proposal.Header.Height, proposal.Header.ChainID, proposal.ProposerIndex)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(proposal)
		return err
	case definition.TendermintPrevote:
		prevote := constypes.NewPrevoteFromBytes(bz)
		if prevote == nil {
			return fmt.Errorf("Prevote Unmarshal Error")
		}

		if err := prevote.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s)received Prevote(height=%d,index=%d)\n", time.Now(), cs.Height, RoundStepString(cs.Step), prevote.Height, prevote.ValidatorIndex)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(prevote)
		return err
	case definition.TendermintPrecommit:
		precommit := constypes.NewPrecommitFromBytes(bz)
		if precommit == nil {
			return fmt.Errorf("Precommit Unmarshal Error")
		}

		if err := precommit.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s)received Precommit(height=%d,index=%d)\n", time.Now(), cs.Height, RoundStepString(cs.Step), precommit.Height, precommit.ValidatorIndex)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(precommit)
		return err
	case definition.TendermintCrossShardProposal:
		csp := constypes.NewCrossShardProposalFromBytes(bz)
		if csp == nil {
			return fmt.Errorf("CrossShardProposal Unmarshal Error")
		}

		if err := csp.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s)received CSP(height=%d,chain=%s)\n", time.Now(), cs.Height, RoundStepString(cs.Step), csp.AggregatedSignatures.Header.Height, csp.AggregatedSignatures.Header.ChainID)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(csp)
		return err
	case definition.TendermintCrossShardAccept:
		if cs.IsIShard() {
			return fmt.Errorf("I-shard received CrossShardAcceptmessage. Is it a problem？")
		}
		csa := constypes.NewCrossShardAcceptFromBytes(bz)
		if csa == nil {
			return fmt.Errorf("CrossShardAccept Unmarshal Error")
		}

		if err := csa.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s)received CSA(height=%d,chain=%s,index=%d)\n", time.Now(), cs.Height, RoundStepString(cs.Step), csa.Height, csa.AcceptChainID, csa.ValidatorIndex)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(csa)
		return err
	case definition.TendermintCrossShardCommit:
		csc := constypes.NewCrossShardCommitFromBytes(bz)
		if csc == nil {
			return fmt.Errorf("CrossShardCommit Unmarshal Error")
		}

		if err := csc.ValidateBasic(); err != nil {
			return err
		}
		cs.stateMtx.Lock()
		fmt.Printf("%v (height=%d,step=%s)received CSC(height=%d,chain=%s)\n", time.Now(), cs.Height, RoundStepString(cs.Step), csc.Height, csc.UpperChainID)
		defer cs.stateMtx.Unlock()
		err := cs.doMessage(csc)
		return err
	default:
		return fmt.Errorf("Tendermint.ConsensusState: Unrecognized message type(" + fmt.Sprint(messageType) + ")")
	}
}

func (cs *ConsensusState) doMessage(msg interface{}) error {
	switch m := msg.(type) {
	case *types.Part:
		if cs.addRelayMessageIfOverTime(m.Height, m.Round, m) {
			return nil
		}
		if err := cs.heightDatas.AddPart(m); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		return cs.handleStateTransition()
	case *constypes.Proposal:
		if cs.addRelayMessageIfOverTime(m.Header.Height, m.Header.Round, m) {
			return nil
		}
		if err := cs.heightDatas.AddProposal(m, cs.calculateProposer()); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	case *constypes.Prevote:
		if cs.addRelayMessageIfOverTime(m.Height, m.Round, m) {
			return nil
		}
		if err := cs.heightDatas.AddPrevote(m); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	case *constypes.Precommit:
		if cs.addRelayMessageIfOverTime(m.Height, m.Round, m) {
			return nil
		}
		if err := cs.heightDatas.AddPrecommit(m); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	case *constypes.CrossShardProposal:
		if cs.addRelayMessageIfOverTime(m.AggregatedSignatures.Header.Height, 0, m) {
			return nil
		}
		if err := cs.heightDatas.AddCrossShardProposal(m); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	case *constypes.CrossShardAccept:
		if cs.addRelayMessageIfOverTime(m.Height, 0, m) {
			return nil
		}
		if err := cs.heightDatas.AddCrossShardAccept(m); err != nil {
			if err == DoNothing {
				log.Println(" Received CSA but DONOTHING")
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	case *constypes.CrossShardCommit:
		if cs.addRelayMessageIfOverTime(m.Height, 0, m) {
			return nil
		}
		if err := cs.heightDatas.AddCrossShardCommit(m); err != nil {
			if err == DoNothing {
				return nil
			}
			return err
		}
		err := cs.handleStateTransition()
		return err
	default:
		return fmt.Errorf("Unkonwn type")
	}
}

func (cs *ConsensusState) handleStateTransition() error {
	switch cs.Step {
	case RoundStepPropose:
		if cs.heightDatas.FinishProposal() {
			cs.WriteLogger("Finish Proposal", false, false)
			cs.doPropose()
			cs.WriteLogger("Enter Prevote", false, false)
		} else {
			return nil
		}
	case RoundStepPrevote:
		if cs.heightDatas.FinishPrevote() {
			cs.WriteLogger("Finish Prevote", false, false)
			cs.doPrevote()
			cs.WriteLogger("Enter Precommit", false, false)
		} else {
			return nil
		}
	case RoundStepPrecommit:
		if cs.heightDatas.FinishPrecommit() {
			cs.WriteLogger("Finish Precommit", false, false)
			cs.doPrecommit()
			cs.WriteLogger("Enter CrossShardProposal/CrossShardAccept", false, false)
		} else {
			return nil
		}
	case RoundStepCrossShardAccept: // RoundStepCrossShardProposal
		if cs.IsIShard() && cs.heightDatas.FinishCrossShardAccept() {
			cs.WriteLogger("Finish CrossShardAccept", false, false)
			cs.doCSA()
			cs.WriteLogger("Enter CrossShardCommit", false, false)
		} else if cs.IsBShard() && cs.heightDatas.FinishCrossShardProposal() {
			cs.WriteLogger("Finish CrossShardProposal", false, false)
			cs.doCSP()
			cs.WriteLogger("Enter CrossShardCommit", false, false)
			log.Printf("CSB:%d,CSC:%d,CSP:%d\n", len(cs.heightDatas.CrossShardBlocks), len(cs.heightDatas.CrossShardCommits), len(cs.heightDatas.CrossShardProposals))
		} else {
			return nil
		}
	case RoundStepCrossShardCommit:
		if cs.IsIShard() && cs.heightDatas.FinishCrossShardCommit_I() {
			cs.WriteLogger("Finish CrossShardCommit", false, false)
			cs.doCSC_I()
		} else if cs.IsBShard() && cs.heightDatas.FinishCrossShardCommit_B() {
			cs.WriteLogger("Finish CrossShardCommit", false, false)
			cs.doCSC_B()
		} else {
			return nil
		}
	}
	return cs.handleStateTransition()
}

func (cs *ConsensusState) calculateProposer() int {
	return (int(cs.Height) + cs.Round) % cs.heightDatas.AllValidatorSet[cs.heightDatas.MyChainID].Size()
}
func (cs *ConsensusState) isProposer() bool {
	return cs.calculateProposer() == cs.signerIndex
}

func (cs *ConsensusState) doPropose() {
	proposalBlock := cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID]
	if cs.IsBShard() && cs.calculateProposer() != cs.signerIndex {

		resp := cs.abci.PreExecutionB(proposalBlock)
		if resp.Code != types.CodeTypeOK || !utils.CompareBytesList(resp.CrossShardDatas, proposalBlock.CrossShardDatas) {

			panic("B shard pre-execution error")
		}
	}
	if !bytes.Equal(cs.LastReceiptRoot, proposalBlock.ReceiptRoot) {
		fmt.Println(cs.LastReceiptRoot)
		fmt.Println(proposalBlock.ReceiptRoot)
		panic("The receipt tree does not match")
	}
	if !bytes.Equal(cs.LastStateRoot, proposalBlock.StateRoot) {
		fmt.Println(cs.LastStateRoot)
		fmt.Println(proposalBlock.StateRoot)
		panic("Data state is inconsistent")
	}
	cs.BlockHash = proposalBlock.Hash()
	vote := constypes.NewPrevote(cs.Height, cs.Round, cs.BlockHash, cs.signerIndex)
	sig, err := cs.signer.SignType(vote)
	if err != nil {
		panic(err)
	}
	vote.Signature = sig
	cs.heightDatas.AddPrevote(vote)
	cs.SendInternal(vote.ProtoBytes(), inter.TendermintPrevote)
	cs.Next()
}

func (cs *ConsensusState) doPrevote() {
	vote := constypes.NewPrecommit(cs.Height, cs.Round, cs.BlockHash, cs.signerIndex)
	sig, err := cs.signer.SignType(vote)
	if err != nil {
		panic(err)
	}
	vote.Signature = sig
	cs.heightDatas.AddPrecommit(vote)
	cs.SendInternal(vote.ProtoBytes(), inter.TendermintPrecommit)
	cs.Next()
}

func (cs *ConsensusState) doPrecommit() {
	var shardType uint32 = constypes.ShardTypeB
	if cs.IsIShard() {
		shardType = constypes.ShardTypeI
	}

	if cs.IsIShard() {
		proposal := constypes.NewCrossShardProposal(cs.heightDatas.GenerateAggregatePrecommitSignature(), shardType)
		cs.heightDatas.CrossShardProposals[cs.heightDatas.MyChainID] = proposal
		if cs.isProposer() {
			go cs.SendToRelatedShards(proposal.ProtoBytes(), inter.TendermintCrossShardProposal)
			part_set, ok := cs.heightDatas.CrossShardPartSets[cs.heightDatas.MyChainID]
			if !ok {
				proposalBlock := cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID]
				part_set = types.PartSetFromBlock(proposalBlock, cs.MaxPartSize, cs.Round)
			}
			for _, part := range part_set.Parts {
				go cs.SendToRelatedShards(part.ProtoBytes(), inter.Part)
			}
		}

	} else {
		proposalBlock := cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID]
		block := &types.Block{
			Header: proposalBlock.Header,
			Body: types.Body{
				BodyTxs: nil,
			},
			CrossShardBody: proposalBlock.CrossShardBody,
			Relation:       proposalBlock.Relation,
		}
		part_set := types.PartSetFromBlock(block, cs.MaxPartSize, cs.Round)
		proposal := constypes.NewCrossShardProposal(cs.heightDatas.GenerateAggregatePrecommitSignature(), shardType)
		proposal.AggregatedSignatures.Header = part_set.Header
		cs.heightDatas.CrossShardProposals[cs.heightDatas.MyChainID] = proposal
		if cs.isProposer() {
			go cs.SendToRelatedShards(proposal.ProtoBytes(), inter.TendermintCrossShardProposal)
			go cs.SendTo2LevelRelatedShards(proposal.ProtoBytes(), inter.TendermintCrossShardProposal)
			for _, part := range part_set.Parts {
				go cs.SendToRelatedShards(part.ProtoBytes(), inter.Part)
				go cs.SendTo2LevelRelatedShards(part.ProtoBytes(), inter.Part)
			}
		}
	}

	cs.Next()
}

func (cs *ConsensusState) doCSA() {
	relatedShards := make([]string, 0, len(cs.heightDatas.ShardInfo.RelatedShards))
	for key := range cs.heightDatas.ShardInfo.RelatedShards {
		relatedShards = append(relatedShards, key)
	}
	cs.sortShards(relatedShards)

	blocks := make([]*types.Block, len(relatedShards))
	for i, shard := range relatedShards {
		blocks[i] = cs.heightDatas.CrossShardBlocks[shard]
	}

	resp := cs.abci.PreExecutionI(blocks)

	for i, shard := range relatedShards {
		csa := constypes.NewCrossShardAccept(
			shard, cs.heightDatas.MyChainID, cs.Height,
			blocks[i].Hash(), resp.CommitStatusVote[i],
			cs.signerIndex,
		)
		if sig, err := cs.signer.SignType(csa); err != nil {
			panic(err)
		} else {
			csa.Signature = sig
		}
		go cs.SendTo(shard, csa.ProtoBytes(), inter.TendermintCrossShardAccept)
	}
	cs.Next()
}

func (cs *ConsensusState) doCSP() {
	csc := cs.heightDatas.GenerateCrossShardCommit()
	cs.heightDatas.CrossShardCommits[cs.heightDatas.MyChainID] = csc
	if cs.isProposer() {
		cs.SendToRelatedShards(csc.ProtoBytes(), inter.TendermintCrossShardCommit)
	}
	cs.Next()
}

func (cs *ConsensusState) doCSC_B() {
	b_neighbours := make([]string, 0, len(cs.heightDatas.ShardInfo.Related2LevelShards)+1)
	b_neighbours = append(b_neighbours, cs.heightDatas.MyChainID)
	i_neighbours := make([]string, 0, len(cs.heightDatas.ShardInfo.RelatedShards)+1)
	i_neighbours = append(i_neighbours, cs.heightDatas.MyChainID)
	for shard := range cs.heightDatas.ShardInfo.Related2LevelShards {
		b_neighbours = append(b_neighbours, shard)
	}
	cs.sortShards(b_neighbours)
	for shard := range cs.heightDatas.ShardInfo.RelatedShards {
		i_neighbours = append(i_neighbours, shard)
	}
	resp, commitStatus := cs.executionAllBlocks(b_neighbours, i_neighbours)

	cs.ChainedHash = make([][]byte, 0, len(i_neighbours)+len(b_neighbours)-2)
	cs.RelatedCommitStatus = make([][]byte, 0, len(b_neighbours)-1)
	for i, bshard := range b_neighbours {
		if bshard == cs.heightDatas.MyChainID {
			cs.CommitStatus = commitStatus[i]
			block := cs.heightDatas.CrossShardBlocks[bshard]
			cs.BlockHash = block.Hash()
			cs.LastBlockTime = block.Time
		} else {
			cs.RelatedCommitStatus = append(cs.RelatedCommitStatus, commitStatus[i])
			cs.ChainedHash = append(cs.ChainedHash, cs.heightDatas.CrossShardBlocks[bshard].Hash())
		}
	}
	for _, ishard := range i_neighbours {
		if ishard != cs.heightDatas.MyChainID {
			cs.ChainedHash = append(cs.ChainedHash, cs.heightDatas.CrossShardBlocks[ishard].Hash())
		}
	}
	inner_receipts := resp.InnerShardResps[cs.heightDatas.MyChainID]
	cross_receipts := resp.CrossShardResps[cs.heightDatas.MyChainID]
	receiptTree := make([][]byte, 0, len(inner_receipts)+len(cross_receipts))
	for _, receipt := range cross_receipts {
		receiptTree = append(receiptTree, receipt.ProtoBytes())
	}
	for _, receipt := range inner_receipts {
		receiptTree = append(receiptTree, receipt.ProtoBytes())
	}
	cs.LastReceiptRoot, _ = merkle.ProofsFromByteSlices(receiptTree)
	cs.LastStateRoot = cs.abci.Commit()

	block := cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID]
	for _, shard := range b_neighbours {
		shardBlock := cs.heightDatas.CrossShardBlocks[shard]
		cs.store.SetBlockByHeight(shardBlock.Height, shardBlock.ChainID, shardBlock)
	}
	for _, shard := range i_neighbours {
		shardBlock := cs.heightDatas.CrossShardBlocks[shard]
		cs.store.SetBlockByHeight(shardBlock.Height, shardBlock.ChainID, shardBlock)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		cs.mempool.Update(cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID].BodyTxs, nil)
	}()
	go func() {
		defer wg.Done()
		cs.cross_shard_mempool.Update(cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID].CrossShardTxs, cs.CommitStatus)
	}()

	var allTxs, passTxs = 0, 0
	innerTxs := block.Body.BodyTxs.Size()
	allTxs = block.CrossShardTxs.Size()
	bv := utils.NewBitArrayFromByte(cs.CommitStatus)
	for i := 0; i < bv.Size(); i++ {
		if bv.GetIndex(i) {
			passTxs++
		}
	}

	cs.WriteLogger(fmt.Sprintf("finish[%d,%d,%d]", innerTxs, passTxs, allTxs), false, true)

	wg.Wait()

	cs.enterNewHeight()
}
func (cs *ConsensusState) doCSC_I() {
	b_neighbours, i_neighbours := make([]string, 0, len(cs.heightDatas.ShardInfo.RelatedShards)), []string{cs.heightDatas.MyChainID}
	for shard := range cs.heightDatas.ShardInfo.RelatedShards {
		b_neighbours = append(b_neighbours, shard)
	}
	cs.sortShards(b_neighbours)
	resp, commitStatus := cs.executionAllBlocks(b_neighbours, i_neighbours)

	cs.ChainedHash = make([][]byte, 0, len(b_neighbours))
	cs.RelatedCommitStatus = make([][]byte, 0, len(b_neighbours))
	for i, bshard := range b_neighbours {
		cs.RelatedCommitStatus = append(cs.RelatedCommitStatus, commitStatus[i])
		cs.ChainedHash = append(cs.ChainedHash, cs.heightDatas.CrossShardBlocks[bshard].Hash())
	}
	cs.CommitStatus = nil
	block := cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID]
	cs.BlockHash = block.Hash()
	cs.LastBlockTime = block.Time

	inner_receipts := resp.InnerShardResps[cs.heightDatas.MyChainID]
	receiptTree := make([][]byte, 0, len(inner_receipts))
	for _, receipt := range inner_receipts {
		receiptTree = append(receiptTree, receipt.ProtoBytes())
	}
	cs.LastReceiptRoot, _ = merkle.ProofsFromByteSlices(receiptTree)
	cs.LastStateRoot = cs.abci.Commit()

	cs.store.SetBlockByHeight(cs.Height, cs.heightDatas.MyChainID, block)
	for _, shard := range b_neighbours {
		shardBlock := cs.heightDatas.CrossShardBlocks[shard]
		cs.store.SetBlockByHeight(shardBlock.Height, shardBlock.ChainID, shardBlock)
	}

	cs.mempool.Update(block.BodyTxs, nil)

	allTxs := block.BodyTxs.Size()

	cs.WriteLogger(fmt.Sprintf("finish[%d,0,0]", allTxs), false, true)

	cs.enterNewHeight()
}

func (cs *ConsensusState) executionAllBlocks(b_neighbours, i_neighbours []string) (*types.ABCIExecutionResponse, [][]byte) {
	b_blocks := make([]*types.Block, len(b_neighbours))
	i_blocks := make([]*types.Block, len(i_neighbours))
	commitStatuses := make([]*utils.BitVector, len(b_neighbours))
	commitStatusBzList := make([][]byte, len(b_neighbours))
	for i, shard := range b_neighbours {
		block := cs.heightDatas.CrossShardBlocks[shard]
		csc := cs.heightDatas.CrossShardCommits[shard]
		var total_commit_status *utils.BitVector = nil
		for _, commit_status := range csc.CommitStatusList {
			//fmt.Println(commit_status)
			if total_commit_status == nil {
				total_commit_status = utils.NewBitArrayFromByte(commit_status)
			} else {
				total_commit_status = total_commit_status.And(utils.NewBitArrayFromByte(commit_status))
			}
		}
		commitStatuses[i] = total_commit_status
		commitStatusBzList[i] = total_commit_status.Byte()
		b_blocks[i] = block
	}
	for i, shard := range i_neighbours {
		i_blocks[i] = cs.heightDatas.CrossShardBlocks[shard]
	}
	return cs.abci.Execution(b_blocks, commitStatuses, i_blocks), commitStatusBzList
}

func (cs *ConsensusState) enterNewHeight() {
	cs.Step = RoundStepCrossShardCommit
	cs.Next()

	cs.WriteLogger("enter New Height", true, false)
	cs.LastBlockHash, cs.LastChainedHash = cs.BlockHash, cs.ChainedHash
	cs.LastCommitStatus, cs.LastRelatedCommitStatus = cs.CommitStatus, cs.RelatedCommitStatus
	cs.heightDatas.NextHeight()

	cs.enterNewRound()
}
func (cs *ConsensusState) enterNewRound() {
	cs.Step = RoundStepPropose
	//fmt.Println(cs.calculateProposer())
	if cs.isProposer() {
		cs.createBlockAndBroadcast()
	}
	if relay_height, ok := cs.relayMessageChannel[cs.Height]; ok {
		if relayRound, ok := relay_height[cs.Round]; ok {
			for _, msg := range relayRound {
				cs.doMessage(msg)
			}
		}
	}
}

func (cs *ConsensusState) createBlockAndBroadcast() {
	relation := types.Relation{
		RelatedCommitStatus: cs.LastRelatedCommitStatus,
		RelatedHash:         cs.LastChainedHash,
		MyCommitStatus:      cs.LastCommitStatus,
	}
	var blockType int8 = 0
	if cs.IsBShard() {
		blockType = types.BLOCKTYPE_B
	} else {
		blockType = types.BLOCKTYPE_I
	}
	Header := types.Header{
		BlockType:   blockType,
		HashPointer: cs.LastBlockHash,
		ChainID:     cs.heightDatas.MyChainID,
		Height:      cs.Height,
		//Time:        time.Now(),

		ReceiptRoot: cs.LastReceiptRoot,
		StateRoot:   cs.LastStateRoot,
	}

	var block *types.Block
	if cs.IsBShard() {
		innerTxs, innerTxsSize, err := cs.mempool.ReapTx(cs.maxBlockTxNum / 2)
		if err != nil {
			panic(err)
		}
		crossShardTxs, _, err := cs.cross_shard_mempool.ReapTx(cs.maxBlockTxNum - innerTxsSize)
		if err != nil {
			panic(err)
		}
		Body := types.Body{BodyTxs: innerTxs}
		CrossShardBody := types.CrossShardBody{CrossShardTxs: crossShardTxs}
		block = &types.Block{
			Header:         Header,
			Relation:       relation,
			Body:           Body,
			CrossShardBody: CrossShardBody,
		}
	} else {
		innerTxs, _, err := cs.mempool.ReapTx(cs.maxBlockTxNum)
		if err != nil {
			panic(err)
		}
		var crossShardTxs types.Txs = nil
		Body := types.Body{BodyTxs: innerTxs}
		CrossShardBody := types.CrossShardBody{CrossShardTxs: crossShardTxs}

		block = &types.Block{
			Header:         Header,
			Relation:       relation,
			Body:           Body,
			CrossShardBody: CrossShardBody,
		}
	}

	interval_dst := cs.LastBlockTime.Add(cs.MinBlockInterval)
	if t := time.Now(); t.Before(interval_dst) {
		block.Header.Time = interval_dst
	} else {
		block.Header.Time = t
	}

	if cs.IsBShard() {
		resp := cs.abci.PreExecutionB(block)
		block.CrossShardBody.CrossShardDatas = resp.CrossShardDatas
	}

	part_set := types.PartSetFromBlock(block, cs.MaxPartSize, cs.Round)
	proposal := constypes.NewProposal(part_set.Header, cs.signerIndex, part_set.BlockHeaderHash)
	if sig, err := cs.signer.SignType(proposal); err != nil {
		panic(err)
	} else {
		proposal.Signature = sig
	}

	log.Println("PartSize", len(part_set.Parts))
	log.Println("Proposal", *proposal)

	cs.heightDatas.CrossShardBlocks[cs.heightDatas.MyChainID] = block
	cs.heightDatas.CrossShardPartSets[cs.heightDatas.MyChainID] = part_set
	cs.heightDatas.Proposal = proposal

	go func() {

		time.Sleep(interval_dst.Sub(time.Now()))

		cs.SendInternal(proposal.ProtoBytes(), inter.TendermintProposal)
		for _, part := range part_set.Parts {
			cs.SendInternal(part.ProtoBytes(), inter.Part)
		}
	}()
}

// ============================================
func (cs *ConsensusState) SendTo(shardID string, bz []byte, messageType uint32) {
	cs.p2p.SendToShard(shardID, p2p.ChannelIDConsensusState, bz, messageType)
	if messageType == inter.TendermintCrossShardAccept {
		log.Printf("broadcast CrossShardAccept to %sshard ", shardID)
	}
}
func (cs *ConsensusState) SendInternal(bz []byte, messageType uint32) {
	cs.p2p.SendToShard(cs.heightDatas.MyChainID, p2p.ChannelIDConsensusState, bz, messageType)
	if messageType == inter.TendermintPrevote {
		log.Printf("intra-shard broadcast Prevote")
	} else if messageType == inter.TendermintPrecommit {
		log.Printf("intra-shard broadcast Precommit")
	}
}
func (cs *ConsensusState) SendToRelatedShards(bz []byte, messageType uint32) {
	relatedShards := cs.heightDatas.ShardInfo.RelatedShards
	for shard := range relatedShards {
		cs.SendTo(shard, bz, messageType)
		if messageType == inter.TendermintCrossShardProposal {
			log.Printf(" broadcast CrossShardProposal to %s shard", shard)
		} else if messageType == inter.Part {
		}
	}
}
func (cs *ConsensusState) SendTo2LevelRelatedShards(bz []byte, messageType uint32) {
	relatedShards := cs.heightDatas.ShardInfo.Related2LevelShards
	for shard := range relatedShards {
		cs.SendTo(shard, bz, messageType)
		if messageType == inter.TendermintCrossShardProposal {
			log.Printf("broadcast CrossShardProposal to %sshard ", shard)
		} else if messageType == inter.Part {
		}
	}
}
func (cs *ConsensusState) SendToRelatedShardsExcept(bz []byte, messageType uint32, exception string) {
	relatedShards := cs.heightDatas.ShardInfo.RelatedShards
	for shard := range relatedShards {
		if shard == exception {
			continue
		}
		cs.SendTo(shard, bz, messageType)
	}
}
