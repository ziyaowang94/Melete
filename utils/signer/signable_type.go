package signer

type SignableType interface {
	SignBytes() []byte
}
