package merkle

import (
	"bytes"
	"errors"
	"fmt"

	tmhash "emulator/crypto/hash"

	pbmerkle "emulator/proto/crypto/merkle"
)

const (
	// MaxAunts is the maximum number of aunts that can be included in a Proof.
	// This corresponds to a tree of size 2^100, which should be sufficient for all conceivable purposes.
	// This maximum helps prevent Denial-of-Service attacks by limitting the size of the proofs.
	MaxAunts = 100
)

// Proof represents a Merkle proof.
// NOTE: The convention for proofs is to include leaf hashes but to
// exclude the root hash.
// This convention is implemented across IAVL range proofs as well.
// Keep this consistent unless there's a very good reason to change
// everything.  This also affects the generalized proof system as
// well.
type Proof struct {
	Total    int64    `json:"total"`     // Total number of items.
	Index    int64    `json:"index"`     // Index of item to prove.
	LeafHash []byte   `json:"leaf_hash"` // Hash of item value.
	Aunts    [][]byte `json:"aunts"`     // Hashes from leaf's sibling to a root's child.
}

// ProofsFromByteSlices computes inclusion proof for given items.
// proofs[0] is the proof for items[0].
func ProofsFromByteSlices(items [][]byte) (rootHash []byte, proofs []*Proof) {
	trails, rootSPN := trailsFromByteSlices(items)
	rootHash = rootSPN.Hash
	proofs = make([]*Proof, len(items))
	for i, trail := range trails {
		proofs[i] = &Proof{
			Total:    int64(len(items)),
			Index:    int64(i),
			LeafHash: trail.Hash,
			Aunts:    trail.FlattenAunts(),
		}
	}
	return
}

// Verify that the Proof proves the root hash.
// Check sp.Index/sp.Total manually if needed
func (sp *Proof) Verify(rootHash []byte, leaf []byte) error {
	leafHash := leafHash(leaf)
	if sp.Total < 0 {
		return errors.New("proof total must be positive")
	}
	if sp.Index < 0 {
		return errors.New("proof index cannot be negative")
	}
	if !bytes.Equal(sp.LeafHash, leafHash) {
		return fmt.Errorf("invalid leaf hash: wanted %X got %X", leafHash, sp.LeafHash)
	}
	computedHash := sp.ComputeRootHash()
	if !bytes.Equal(computedHash, rootHash) {
		return fmt.Errorf("invalid root hash: wanted %X got %X", rootHash, computedHash)
	}
	return nil
}

// Compute the root hash given a leaf hash.  Does not verify the result.
func (sp *Proof) ComputeRootHash() []byte {
	return computeHashFromAunts(
		sp.Index,
		sp.Total,
		sp.LeafHash,
		sp.Aunts,
	)
}

// String implements the stringer interface for Proof.
// It is a wrapper around StringIndented.
func (sp *Proof) String() string {
	return sp.StringIndented("")
}

// StringIndented generates a canonical string representation of a Proof.
func (sp *Proof) StringIndented(indent string) string {
	return fmt.Sprintf(`Proof{
%s  Aunts: %X
%s}`,
		indent, sp.Aunts,
		indent)
}

// ValidateBasic performs basic validation.
// NOTE: it expects the LeafHash and the elements of Aunts to be of size tmhash.Size,
// and it expects at most MaxAunts elements in Aunts.
func (sp *Proof) ValidateBasic() error {
	if sp.Total < 0 {
		return errors.New("negative Total")
	}
	if sp.Index < 0 {
		return errors.New("negative Index")
	}
	if len(sp.LeafHash) != tmhash.Size {
		return fmt.Errorf("expected LeafHash size to be %d, got %d", tmhash.Size, len(sp.LeafHash))
	}
	if len(sp.Aunts) > MaxAunts {
		return fmt.Errorf("expected no more than %d aunts, got %d", MaxAunts, len(sp.Aunts))
	}
	for i, auntHash := range sp.Aunts {
		if len(auntHash) != tmhash.Size {
			return fmt.Errorf("expected Aunts#%d size to be %d, got %d", i, tmhash.Size, len(auntHash))
		}
	}
	return nil
}

// ToProto converts a Proof struct to its protobuf representation.
func (p *Proof) ToProto() *pbmerkle.Proof {
	return &pbmerkle.Proof{
		Total:    p.Total,
		Index:    p.Index,
		LeafHash: p.LeafHash,
		Aunts:    p.Aunts,
	}
}

// NewProof creates a new Proof struct from the given protobuf representation.
func NewProofFromProto(proto *pbmerkle.Proof) *Proof {
	return &Proof{
		Total:    proto.Total,
		Index:    proto.Index,
		LeafHash: proto.LeafHash,
		Aunts:    proto.Aunts,
	}
}

// Use the leafHash and innerHashes to get the root merkle hash.
// If the length of the innerHashes slice isn't exactly correct, the result is nil.
// Recursive impl.
func computeHashFromAunts(index, total int64, leafHash []byte, innerHashes [][]byte) []byte {
	if index >= total || index < 0 || total <= 0 {
		return nil
	}
	switch total {
	case 0:
		panic("Cannot call computeHashFromAunts() with 0 total")
	case 1:
		if len(innerHashes) != 0 {
			return nil
		}
		return leafHash
	default:
		if len(innerHashes) == 0 {
			return nil
		}
		numLeft := getSplitPoint(total)
		if index < numLeft {
			leftHash := computeHashFromAunts(index, numLeft, leafHash, innerHashes[:len(innerHashes)-1])
			if leftHash == nil {
				return nil
			}
			return innerHash(leftHash, innerHashes[len(innerHashes)-1])
		}
		rightHash := computeHashFromAunts(index-numLeft, total-numLeft, leafHash, innerHashes[:len(innerHashes)-1])
		if rightHash == nil {
			return nil
		}
		return innerHash(innerHashes[len(innerHashes)-1], rightHash)
	}
}

// ProofNode is a helper structure to construct merkle proof.
// The node and the tree is thrown away afterwards.
// Exactly one of node.Left and node.Right is nil, unless node is the root, in which case both are nil.
// node.Parent.Hash = hash(node.Hash, node.Right.Hash) or
// hash(node.Left.Hash, node.Hash), depending on whether node is a left/right child.
type ProofNode struct {
	Hash   []byte
	Parent *ProofNode
	Left   *ProofNode // Left sibling  (only one of Left,Right is set)
	Right  *ProofNode // Right sibling (only one of Left,Right is set)
}

// FlattenAunts will return the inner hashes for the item corresponding to the leaf,
// starting from a leaf ProofNode.
func (spn *ProofNode) FlattenAunts() [][]byte {
	// Nonrecursive impl.
	innerHashes := [][]byte{}
	for spn != nil {
		switch {
		case spn.Left != nil:
			innerHashes = append(innerHashes, spn.Left.Hash)
		case spn.Right != nil:
			innerHashes = append(innerHashes, spn.Right.Hash)
		default:
			break
		}
		spn = spn.Parent
	}
	return innerHashes
}

// trails[0].Hash is the leaf hash for items[0].
// trails[i].Parent.Parent....Parent == root for all i.
func trailsFromByteSlices(items [][]byte) (trails []*ProofNode, root *ProofNode) {
	// Recursive impl.
	switch len(items) {
	case 0:
		return []*ProofNode{}, &ProofNode{emptyHash(), nil, nil, nil}
	case 1:
		trail := &ProofNode{leafHash(items[0]), nil, nil, nil}
		return []*ProofNode{trail}, trail
	default:
		k := getSplitPoint(int64(len(items)))
		lefts, leftRoot := trailsFromByteSlices(items[:k])
		rights, rightRoot := trailsFromByteSlices(items[k:])
		rootHash := innerHash(leftRoot.Hash, rightRoot.Hash)
		root := &ProofNode{rootHash, nil, nil, nil}
		leftRoot.Parent = root
		leftRoot.Right = rightRoot
		rightRoot.Parent = root
		rightRoot.Left = leftRoot
		return append(lefts, rights...), root
	}
}
