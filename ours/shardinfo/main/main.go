package main

import (
	"fmt"

	"github.com/herumi/bls-eth-go-binary/bls"
)

func main() {
	// Initialize BLS library
	err := bls.Init(bls.BLS12_381)
	if err != nil {
		fmt.Println("Error initializing BLS library:", err)
		return
	}

	// Generate three random private keys
	privateKey1 := bls.SecretKey{}
	privateKey2 := bls.SecretKey{}
	privateKey3 := bls.SecretKey{}

	privateKey1.SetByCSPRNG()
	privateKey2.SetByCSPRNG()
	privateKey3.SetByCSPRNG()

	// Derive public keys from private keys
	publicKey4 := privateKey1.GetPublicKey()
	publicKey2 := privateKey2.GetPublicKey()
	publicKey3 := privateKey3.GetPublicKey()

	bz := publicKey4.SerializeToHexStr()
	fmt.Println(bz)
	publicKey1 := new(bls.PublicKey)
	err = publicKey1.DeserializeHexStr(bz)
	if err != nil {
		panic(err)
	}

	// Create a message to be signed
	message := "Hello, BLS!"

	// Sign the message using each private key
	signature1 := privateKey1.Sign(message)
	signature2 := privateKey2.Sign(message)
	signature3 := privateKey3.Sign(message)

	// Aggregate the signatures
	aggSignature := bls.Sign{}
	aggSignature.Add(signature1)
	aggSignature.Add(signature2)
	aggSignature.Add(signature3)

	// Verify the aggregated signature with the aggregated public key
	aggPublicKey := publicKey1
	aggPublicKey.Add(publicKey2)
	aggPublicKey.Add(publicKey3)

	if aggSignature.Verify(aggPublicKey, message) {
		fmt.Println("Aggregated signature verified successfully!")
	} else {
		fmt.Println("Invalid aggregated signature")
	}
}
