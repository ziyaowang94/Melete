package constypes

import (
	"emulator/crypto/hash"
	"emulator/ours/types"
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	proposal := NewProposal(
		&types.PartSetHeader{
			Height:  10,
			Round:   10,
			Total:   10,
			Root:    hash.Sum([]byte("111")),
			ChainID: "a",
		},
		10,
		hash.Sum([]byte("111")),
	)
	fmt.Println(proposal.ValidateBasic())
	bz := proposal.ProtoBytes()
	p := NewProposalFromBytes(bz)
	fmt.Println(p)

}
