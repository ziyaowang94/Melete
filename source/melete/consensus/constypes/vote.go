package constypes

import (
	"emulator/melete/types"
	pbcons "emulator/proto/melete/consensus/types"
	"emulator/utils"
	"emulator/utils/p2p"
	"emulator/utils/signer"
	"encoding/hex"
	"fmt"

	"emulator/crypto/hash"

	"google.golang.org/protobuf/proto"
)

const (
	VoteTypeNil = iota
	VoteTypeProposal
	VoteTypePrevote
	// VoteTypePrevoteAggregated
	VoteTypePrecommit
	VoteTypePrecommitAggregated
	VoteTypeCrossShardProposal
	VoteTypeCrossShardAccept
	VoteTypeCrossShardCommit

	ShardTypeI
	ShardTypeC
)

// ============ Proposal ========================
type Proposal struct {
	VoteType uint32 // VoteTypeProposal

	Header          *types.PartSetHeader
	ProposerIndex   int
	BlockHeaderHash []byte

	Signature string
}

func NewProposal(header *types.PartSetHeader, proposer_index int, blockHeaderHash []byte) *Proposal {
	return &Proposal{
		VoteType:        VoteTypeProposal,
		Header:          header,
		ProposerIndex:   proposer_index,
		BlockHeaderHash: blockHeaderHash,
	}
}

func (p *Proposal) ToProto() *pbcons.Proposal {
	return &pbcons.Proposal{
		VoteType:        p.VoteType,
		Header:          p.Header.ToProto(),
		ProposerIndex:   int32(p.ProposerIndex),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
	}
}

func (p *Proposal) ProtoBytes() []byte {
	return types.MustProtoBytes(p.ToProto())
}
func (p *Proposal) SignBytes() []byte {
	return types.MustProtoBytes(
		&pbcons.Proposal{
			VoteType:        p.VoteType,
			Header:          p.Header.ToProto(),
			ProposerIndex:   int32(p.ProposerIndex),
			BlockHeaderHash: p.BlockHeaderHash,
			Signature:       "",
		},
	)
}
func NewProposalFromProto(p *pbcons.Proposal) *Proposal {
	return &Proposal{
		VoteType:        p.VoteType,
		Header:          types.NewPartSetHeaderFromProto(p.Header),
		ProposerIndex:   int(p.ProposerIndex),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
	}
}
func NewProposalFromBytes(bz []byte) *Proposal {
	var p = new(pbcons.Proposal)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewProposalFromProto(p)
	}
}
func (p *Proposal) ValidateBasic() error {
	if p.VoteType != VoteTypeProposal {
		return fmt.Errorf(fmt.Sprintf("VoteTypeError: %d is note a VoteType for Proposal", p.VoteType))
	}
	if p.ProposerIndex < 0 {
		return fmt.Errorf(fmt.Sprintf("ProposerIndex is negative: %d", p.ProposerIndex))
	}
	if len(p.BlockHeaderHash) != hash.HashSize {
		return fmt.Errorf(hex.EncodeToString(p.BlockHeaderHash) + " It's not a hash")
	}

	return p.Header.ValidateBasic()
}

// ============ Prevote =================
type Prevote struct {
	VoteType uint32

	Code   int8
	Height int64
	Round  int
	// TimeStamp       time.Time
	BlockHeaderHash []byte

	Signature      string
	ValidatorIndex int
}

func (p *Prevote) IsOK() bool { return p.Code == types.CodeTypeOK }
func (p *Prevote) SetOK(ok bool) {
	if ok {
		p.Code = types.CodeTypeOK
	} else {
		p.Code = types.CodeTypeAbort
	}
}
func (p *Prevote) ValidateBasic() error {
	if p.VoteType != VoteTypePrevote {
		return fmt.Errorf(fmt.Sprintf("VoteTypeError: %d is not a VoteType for Prevote", p.VoteType))
	}
	if len(p.BlockHeaderHash) != hash.HashSize {
		return fmt.Errorf(fmt.Sprintf("Prevote malformed: BlockHeaderHash is not a hash value"))
	}

	return nil
}

func NewPrevote(height int64, round int, blockHeaderHash []byte, validatorIndex int) *Prevote {
	return &Prevote{
		Code:            types.CodeTypeOK,
		VoteType:        VoteTypePrevote,
		Height:          height,
		Round:           round,
		BlockHeaderHash: blockHeaderHash,
		ValidatorIndex:  validatorIndex,
	}
}

func (p *Prevote) ToProto() *pbcons.Prevote {
	return &pbcons.Prevote{
		Code:            int32(p.Code),
		VoteType:        p.VoteType,
		Height:          p.Height,
		Round:           int32(p.Round),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
		ValidatorIndex:  int32(p.ValidatorIndex),
	}
}

func (p *Prevote) ProtoBytes() []byte {
	return types.MustProtoBytes(p.ToProto())
}

func (p *Prevote) SignBytes() []byte {
	return types.MustProtoBytes(
		&pbcons.Prevote{
			Code:            int32(p.Code),
			VoteType:        p.VoteType,
			Height:          p.Height,
			Round:           int32(p.Round),
			BlockHeaderHash: p.BlockHeaderHash,
			Signature:       "",
			ValidatorIndex:  -1,
		},
	)
}

func NewPrevoteFromProto(p *pbcons.Prevote) *Prevote {
	return &Prevote{
		Code:            int8(p.Code),
		VoteType:        p.VoteType,
		Height:          p.Height,
		Round:           int(p.Round),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
		ValidatorIndex:  int(p.ValidatorIndex),
	}
}

func NewPrevoteFromBytes(bz []byte) *Prevote {
	var p = new(pbcons.Prevote)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewPrevoteFromProto(p)
	}
}

// ============ Precommit ====================
type Precommit struct {
	VoteType uint32
	Code     int8

	Height int64
	Round  int
	// TimeStamp       time.Time
	BlockHeaderHash []byte

	Signature      string
	ValidatorIndex int
}

func (p *Precommit) IsOK() bool { return p.Code == types.CodeTypeOK }
func (p *Precommit) SetOK(ok bool) {
	if ok {
		p.Code = types.CodeTypeOK
	} else {
		p.Code = types.CodeTypeAbort
	}
}
func (p *Precommit) ValidateBasic() error {
	if p.VoteType != VoteTypePrecommit {
		return fmt.Errorf(fmt.Sprintf("VoteTypeError: %d is not a Precommit VoteType", p.VoteType))
	}
	if len(p.BlockHeaderHash) != hash.HashSize {
		return fmt.Errorf(fmt.Sprintf("Precommit malformed: BlockHeaderHash is not a hash value"))
	}

	return nil
}

func NewPrecommit(height int64, round int, blockHeaderHash []byte, validatorIndex int) *Precommit {
	return &Precommit{
		Code:            types.CodeTypeOK,
		VoteType:        VoteTypePrecommit,
		Height:          height,
		Round:           round,
		BlockHeaderHash: blockHeaderHash,
		ValidatorIndex:  validatorIndex,
	}
}

func (p *Precommit) ToProto() *pbcons.Precommit {
	return &pbcons.Precommit{
		Code:            int32(p.Code),
		VoteType:        p.VoteType,
		Height:          p.Height,
		Round:           int32(p.Round),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
		ValidatorIndex:  int32(p.ValidatorIndex),
	}
}

func (p *Precommit) ProtoBytes() []byte {
	return types.MustProtoBytes(p.ToProto())
}

func (p *Precommit) SignBytes() []byte {
	return types.MustProtoBytes(
		&pbcons.Precommit{
			Code:            int32(p.Code),
			VoteType:        p.VoteType,
			Height:          p.Height,
			Round:           int32(p.Round),
			BlockHeaderHash: p.BlockHeaderHash,
			Signature:       "",
			ValidatorIndex:  -1,
		},
	)
}

func NewPrecommitFromProto(p *pbcons.Precommit) *Precommit {
	return &Precommit{
		Code:            int8(p.Code),
		VoteType:        p.VoteType,
		Height:          p.Height,
		Round:           int(p.Round),
		BlockHeaderHash: p.BlockHeaderHash,
		Signature:       p.Signature,
		ValidatorIndex:  int(p.ValidatorIndex),
	}
}

func NewPrecommitFromBytes(bz []byte) *Precommit {
	var p = new(pbcons.Precommit)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewPrecommitFromProto(p)
	}
}

// ========== PrecommitAggregated =======================
type PrecommitAggregated struct {
	VoteType uint32

	Code int8

	ValidatorBitVector *utils.BitVector

	Header          *types.PartSetHeader
	BlockHeaderHash []byte
	// TimeStamps []time.Time

	AggregatedSignatures string
}

func (p *PrecommitAggregated) ValidateBasic() error {
	if p.VoteType != VoteTypePrecommitAggregated {
		return fmt.Errorf(fmt.Sprintf("VoteTypeError: %d is not a PrecommitAggregated VoteType", p.VoteType))
	}
	if p.ValidatorBitVector == nil {
		return fmt.Errorf(fmt.Sprintf("PrecommitAggregated BitVector must not be empty"))
	}
	if err := p.Header.ValidateBasic(); err != nil {
		return err
	}
	if len(p.BlockHeaderHash) != hash.HashSize {
		return fmt.Errorf(fmt.Sprintf("Precommit malformed: BlockHeaderHash is not a hash value"))
	}
	return nil
}
func NewPrecommitAggregated(validatorBitVector []byte, header *types.PartSetHeader, blockHeaderHash []byte, aggregatedSignatures string) *PrecommitAggregated {
	return &PrecommitAggregated{
		VoteType:             VoteTypePrecommitAggregated,
		Code:                 types.CodeTypeOK,
		ValidatorBitVector:   utils.NewBitArrayFromByte(validatorBitVector),
		Header:               header,
		BlockHeaderHash:      blockHeaderHash,
		AggregatedSignatures: aggregatedSignatures,
	}
}

func (p *PrecommitAggregated) ToProto() *pbcons.PrecommitAggregated {
	return &pbcons.PrecommitAggregated{
		VoteType:             p.VoteType,
		Code:                 int32(p.Code),
		ValidatorBitVector:   p.ValidatorBitVector.Byte(),
		Header:               p.Header.ToProto(),
		BlockHeaderHash:      p.BlockHeaderHash,
		AggregatedSignatures: p.AggregatedSignatures,
	}
}

func (p *PrecommitAggregated) ProtoBytes() []byte {
	return types.MustProtoBytes(p.ToProto())
}

func NewPrecommitAggregatedFromProto(p *pbcons.PrecommitAggregated) *PrecommitAggregated {
	return &PrecommitAggregated{
		VoteType:             p.VoteType,
		Code:                 int8(p.Code),
		ValidatorBitVector:   utils.NewBitArrayFromByte(p.ValidatorBitVector),
		Header:               types.NewPartSetHeaderFromProto(p.Header),
		BlockHeaderHash:      p.BlockHeaderHash,
		AggregatedSignatures: p.AggregatedSignatures,
	}
}

func NewPrecommitAggregatedFromBytes(bz []byte) *PrecommitAggregated {
	var p = new(pbcons.PrecommitAggregated)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewPrecommitAggregatedFromProto(p)
	}
}

func (p *PrecommitAggregated) Precommit() *Precommit {
	precommit := NewPrecommit(p.Header.Height, p.Header.Round, p.BlockHeaderHash, -1)
	precommit.Code = p.Code
	return precommit
}

func (p *PrecommitAggregated) VerifySignatures(verifier *signer.Verifier, bitVectorBytes []byte) bool {
	if verifier == nil {
		return false
	}
	precommit := p.Precommit()
	return verifier.VerifyAggregateSignature(p.AggregatedSignatures, precommit.SignBytes(), bitVectorBytes)
}
func (p *PrecommitAggregated) IsOK() bool { return p.Code == types.CodeTypeOK }

// ======== CrossShardProposal ================
type CrossShardProposal struct {
	VoteType  uint32
	ShardType uint32

	AggregatedSignatures *PrecommitAggregated
}

func (c *CrossShardProposal) ValidateBasic() error {
	if c.VoteType != VoteTypeCrossShardProposal {
		return fmt.Errorf(fmt.Sprintf("VoteTypeError: %d is not a VoteType for the CrossShardProposal", c.VoteType))
	}
	if c.ShardType != ShardTypeC && c.ShardType != ShardTypeI {
		return fmt.Errorf(fmt.Sprintf("ShardTypeError: %d is not a ShardType for CrossShardProposal"))
	}
	if err := c.AggregatedSignatures.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func NewCrossShardProposal(aggregatedSignatures *PrecommitAggregated, shard_type uint32) *CrossShardProposal {
	return &CrossShardProposal{
		VoteType:             VoteTypeCrossShardProposal,
		AggregatedSignatures: aggregatedSignatures,
		ShardType:            shard_type,
	}
}

func (c *CrossShardProposal) ToProto() *pbcons.CrossShardProposal {
	return &pbcons.CrossShardProposal{
		VoteType:             c.VoteType,
		AggregatedSignatures: c.AggregatedSignatures.ToProto(),
		ShardType:            c.ShardType,
	}
}

func (c *CrossShardProposal) ProtoBytes() []byte {
	return types.MustProtoBytes(c.ToProto())
}

func NewCrossShardProposalFromProto(p *pbcons.CrossShardProposal) *CrossShardProposal {
	return &CrossShardProposal{
		VoteType:             p.VoteType,
		AggregatedSignatures: NewPrecommitAggregatedFromProto(p.AggregatedSignatures),
		ShardType:            p.ShardType,
	}
}

func NewCrossShardProposalFromBytes(bz []byte) *CrossShardProposal {
	var p = new(pbcons.CrossShardProposal)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewCrossShardProposalFromProto(p)
	}
}

// =========== CrossShardAccept ===============
type CrossShardAccept struct {
	VoteType uint32

	UpperChainID  string
	AcceptChainID string

	Height          int64
	BlockHeaderHash []byte

	CommitStatus []byte

	Signature      string
	ValidatorIndex int
}

func (csa *CrossShardAccept) ValidateBasic() error {

	return nil
}

func NewCrossShardAccept(upperChainID, acceptChainID string, height int64, blockHeaderHash []byte, commitStatus []byte, validatorIndex int) *CrossShardAccept {
	return &CrossShardAccept{
		VoteType:        VoteTypeCrossShardAccept,
		UpperChainID:    upperChainID,
		AcceptChainID:   acceptChainID,
		Height:          height,
		BlockHeaderHash: blockHeaderHash,
		CommitStatus:    commitStatus,
		ValidatorIndex:  validatorIndex,
	}
}

func (c *CrossShardAccept) ToProto() *pbcons.CrossShardAccept {
	return &pbcons.CrossShardAccept{
		VoteType:        c.VoteType,
		UpperChainID:    c.UpperChainID,
		AcceptChainID:   c.AcceptChainID,
		Height:          c.Height,
		BlockHeaderHash: c.BlockHeaderHash,
		CommitStatus:    c.CommitStatus,
		Signature:       c.Signature,
		ValidatorIndex:  int32(c.ValidatorIndex),
	}
}

func (c *CrossShardAccept) ProtoBytes() []byte {
	return types.MustProtoBytes(c.ToProto())
}

func (c *CrossShardAccept) SignBytes() []byte {
	return types.MustProtoBytes(
		&pbcons.CrossShardAccept{
			VoteType:        c.VoteType,
			UpperChainID:    c.UpperChainID,
			AcceptChainID:   c.AcceptChainID,
			Height:          c.Height,
			BlockHeaderHash: c.BlockHeaderHash,
			CommitStatus:    c.CommitStatus,
			Signature:       "",
			ValidatorIndex:  -1,
		},
	)
}

func NewCrossShardAcceptFromProto(p *pbcons.CrossShardAccept) *CrossShardAccept {
	return &CrossShardAccept{
		VoteType:        p.VoteType,
		UpperChainID:    p.UpperChainID,
		AcceptChainID:   p.AcceptChainID,
		Height:          p.Height,
		BlockHeaderHash: p.BlockHeaderHash,
		CommitStatus:    p.CommitStatus,
		Signature:       p.Signature,
		ValidatorIndex:  int(p.ValidatorIndex),
	}
}

func NewCrossShardAcceptFromBytes(bz []byte) *CrossShardAccept {
	var p = new(pbcons.CrossShardAccept)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewCrossShardAcceptFromProto(p)
	}
}

// ========= CrossShardCommit =================
type CrossShardCommit struct {
	VoteType uint32

	UpperChainID    string
	Height          int64
	BlockHeaderHash []byte

	ChainList               []string
	AggregatedSignatureList []string
	ValidatorBitVectorList  []*utils.BitVector
	CommitStatusList        [][]byte
}

func (csc *CrossShardCommit) ValidateBasic() error {

	return nil
}

func (csc *CrossShardCommit) getCrossShardAccept(i int, valIndex int) *CrossShardAccept {
	return NewCrossShardAccept(csc.UpperChainID,
		csc.ChainList[i],
		csc.Height,
		csc.BlockHeaderHash,
		csc.CommitStatusList[i],
		valIndex)
}

func NewCrossShardCommit(upperChainID string, height int64, blockHeaderHash []byte, chainList []string, aggregatedSignatureList []string, validatorBitVectorList [][]byte, commitStatusList [][]byte) *CrossShardCommit {
	bvList := make([]*utils.BitVector, len(validatorBitVectorList))
	for i, bz := range validatorBitVectorList {
		bvList[i] = utils.NewBitArrayFromByte(bz)
	}
	return &CrossShardCommit{
		VoteType:                VoteTypeCrossShardCommit,
		UpperChainID:            upperChainID,
		Height:                  height,
		BlockHeaderHash:         blockHeaderHash,
		ChainList:               chainList,
		AggregatedSignatureList: aggregatedSignatureList,
		ValidatorBitVectorList:  bvList,
		CommitStatusList:        commitStatusList,
	}
}

func (c *CrossShardCommit) ToProto() *pbcons.CrossShardCommit {
	protoAggregatedSignatures := make([]string, len(c.AggregatedSignatureList))
	for i, sig := range c.AggregatedSignatureList {
		protoAggregatedSignatures[i] = sig
	}

	protoValidatorBitVectors := make([][]byte, len(c.ValidatorBitVectorList))
	for i, vbv := range c.ValidatorBitVectorList {
		protoValidatorBitVectors[i] = vbv.Byte()
	}

	protoCommitStatusList := make([][]byte, len(c.CommitStatusList))
	for i, status := range c.CommitStatusList {
		protoCommitStatusList[i] = status
	}

	return &pbcons.CrossShardCommit{
		VoteType:                c.VoteType,
		UpperChainID:            c.UpperChainID,
		Height:                  c.Height,
		BlockHeaderHash:         c.BlockHeaderHash,
		ChainList:               c.ChainList,
		AggregatedSignatureList: protoAggregatedSignatures,
		ValidatorBitVectorList:  protoValidatorBitVectors,
		CommitStatusList:        protoCommitStatusList,
	}
}

func (c *CrossShardCommit) ProtoBytes() []byte {
	return types.MustProtoBytes(c.ToProto())
}

func NewCrossShardCommitFromProto(p *pbcons.CrossShardCommit) *CrossShardCommit {
	bvList := make([]*utils.BitVector, len(p.ValidatorBitVectorList))
	for i, bz := range p.ValidatorBitVectorList {
		bvList[i] = utils.NewBitArrayFromByte(bz)
	}
	return &CrossShardCommit{
		VoteType:                p.VoteType,
		UpperChainID:            p.UpperChainID,
		Height:                  p.Height,
		BlockHeaderHash:         p.BlockHeaderHash,
		ChainList:               p.ChainList,
		AggregatedSignatureList: p.AggregatedSignatureList,
		ValidatorBitVectorList:  bvList,
		CommitStatusList:        p.CommitStatusList,
	}
}

func NewCrossShardCommitFromBytes(bz []byte) *CrossShardCommit {
	var p = new(pbcons.CrossShardCommit)
	if err := proto.Unmarshal(bz, p); err != nil {
		return nil
	} else {
		return NewCrossShardCommitFromProto(p)
	}
}

func (csc *CrossShardCommit) ValidateVote(peerList []*p2p.Peer, verifier *signer.Verifier, index int, shardVote int32) error {
	bv := csc.ValidatorBitVectorList[index]
	if len(peerList) == 0 {
		return fmt.Errorf("Nil PeerList while CrossShardCommit")
	}
	if len(peerList) != verifier.Size() {
		return fmt.Errorf("The peerList and ValidatorSet of errors are provided")
	}
	if len(peerList) != bv.Size() {
		return fmt.Errorf("PeerList does not match BitVector length while CrossShardCommit")
	}
	voteTotal := int32(0)
	for i, peer := range peerList {
		if bv.GetIndex(i) {
			voteTotal += peer.Vote
		}
	}
	if !(voteTotal > shardVote*2/3) {
		return fmt.Errorf(fmt.Sprintf("CrossShardCommit has less than 2/3 votes for shard %s", csc.ChainList[index]))
	}
	msg := csc.getCrossShardAccept(index, 0)
	if !verifier.VerifyAggregateSignature(csc.AggregatedSignatureList[index], msg.SignBytes(), bv.Byte()) {
		return fmt.Errorf(fmt.Sprintf("There is a problem with the aggregate signature of the shard %s of CrossShardCommit", csc.ChainList[index]))
	}
	return nil
}

func isNotHash(bz []byte) bool { return len(bz) != hash.HashSize }
