package constypes

import (
	"bytes"
	"emulator/utils/p2p"
	crypto "emulator/utils/signer"
	"fmt"
)

type (
	CSAShardSet struct {
		ValSet            *crypto.Verifier
		PeerList          []*p2p.Peer
		TotalVotingPower  int32
		Height            int64
		ChainID           string
		SigMap            map[string][]*CrossShardAccept
		KeyVotePower      map[string]int32
		IndexSeen         map[int][]byte
		HasMaj23Flag      bool
		Maj23Key          string
		Maj23CommitStatus []byte
	}
)

func NewCSAShardSet(h int64, chain_id string, peerList []*p2p.Peer, valSet *crypto.Verifier) *CSAShardSet {
	var totalVotingPower int32 = 0
	for _, peer := range peerList {
		totalVotingPower += peer.Vote
	}
	return &CSAShardSet{
		ValSet:            valSet,
		PeerList:          peerList,
		TotalVotingPower:  totalVotingPower,
		Height:            h,
		ChainID:           chain_id,
		SigMap:            make(map[string][]*CrossShardAccept),
		KeyVotePower:      make(map[string]int32),
		IndexSeen:         make(map[int][]byte),
		HasMaj23Flag:      false,
		Maj23Key:          "",
		Maj23CommitStatus: nil,
	}
}

func (s *CSAShardSet) AddCSA(csa *CrossShardAccept) error {
	if u, ok := s.IndexSeen[csa.ValidatorIndex]; ok {
		if bytes.Equal(u, csa.BlockHeaderHash) {
			return nil
		}
		return fmt.Errorf("Repeat CrorssShardAccept to different blocks, Byzantine nodes")
	}
	if csa.Height != s.Height {
		return fmt.Errorf("CrossShardAccept include invalid height: want %d, get %d", s.Height, csa.Height)
	}
	if csa.AcceptChainID != s.ChainID {
		return fmt.Errorf("CrossShardAccept include invalid chain_id: want %s, get %s", s.ChainID, csa.AcceptChainID)
	}
	if csa.ValidatorIndex < 0 || csa.ValidatorIndex >= s.ValSet.Size() {
		return fmt.Errorf("CrossShardAccept invalid ValidatorIndex")
	}
	index := csa.ValidatorIndex
	if !s.ValSet.Verify(csa.Signature, csa.SignBytes(), index) {
		return fmt.Errorf("CrossShardAccept Signature Error")
	}
	csaKey := string(csa.CommitStatus)
	s.SigMap[csaKey] = append(s.SigMap[csaKey], csa)
	s.IndexSeen[csa.ValidatorIndex] = csa.BlockHeaderHash
	s.KeyVotePower[csaKey] += s.PeerList[index].Vote
	if s.KeyVotePower[csaKey] > s.TotalVotingPower*2/3 {
		if s.HasMaj23() && s.Maj23Key != csaKey {
			return fmt.Errorf("There are two 2f+1 voting sets, there is ambiguity, please deal with it in time")
		}
		s.HasMaj23Flag = true
		s.Maj23Key = csaKey
		s.Maj23CommitStatus = csa.CommitStatus
	}
	return nil
}

func (s *CSAShardSet) HasMaj23() bool {
	return s.HasMaj23Flag
}

func (s *CSAShardSet) GetMaj23() ([]byte, []*CrossShardAccept) {
	if !s.HasMaj23() {
		return nil, nil
	}
	return s.Maj23CommitStatus, s.SigMap[s.Maj23Key]
}
