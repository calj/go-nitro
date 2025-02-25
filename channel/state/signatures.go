package state

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/statechannels/go-nitro/types"
)

// Signature is an ECDSA signature
type Signature struct {
	R []byte
	S []byte
	V byte
}

// SignEthereumMessage accepts an arbitrary message, prepends a known message,
// hashes the result using keccak256 and calculates the secp256k1 signature
// of the hash using the provided secret key. The known message added to the input before hashing is
// "\x19Ethereum Signed Message:\n" + len(message).
// See https://github.com/ethereum/go-ethereum/pull/2940 and EIPs 191, 721.
func SignEthereumMessage(message []byte, secretKey []byte) (Signature, error) {
	digest := computeEthereumSignedMessageDigest(message)
	concatenatedSignature, error := secp256k1.Sign(digest, secretKey)
	if error != nil {
		return Signature{}, error
	}
	return splitSignature(concatenatedSignature), nil
}

// RecoverEthereumMessageSigner accepts a message (bytestring) and signature generated by SignEthereumMessage.
// It reconstructs the appropriate digest and recovers an address via secp256k1 public key recovery
func RecoverEthereumMessageSigner(message []byte, signature Signature) (common.Address, error) {
	digest := computeEthereumSignedMessageDigest(message)
	pubKey, error := secp256k1.RecoverPubkey(digest, joinSignature(signature))
	if error != nil {
		return types.Address{}, error
	}
	ecdsaPubKey, error := crypto.UnmarshalPubkey(pubKey)
	if error != nil {
		return types.Address{}, error
	}
	crypto.PubkeyToAddress(*ecdsaPubKey)
	return crypto.PubkeyToAddress(*ecdsaPubKey), error
}

// computeEthereumSignedMessageDigest accepts an arbitrary message, prepends a known message,
// and hashes the result using keccak256. The known message added to the input before hashing is
// "\x19Ethereum Signed Message:\n" + len(message).
func computeEthereumSignedMessageDigest(message []byte) []byte {
	return crypto.Keccak256(
		[]byte(
			fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), string(message)),
		),
	)
}

// splitSignature takes a 65 bytes signature in the [R||S||V] format and returns the individual components
func splitSignature(concatenatedSignature []byte) (signature Signature) {
	signature.R = concatenatedSignature[:32]
	signature.S = concatenatedSignature[32:64]
	signature.V = concatenatedSignature[64]
	return
}

// joinSignature takes a Signature and returns a 65 byte concatenatedSignature in the [R||S||V] format
func joinSignature(signature Signature) (concatenatedSignature []byte) {
	concatenatedSignature = append(concatenatedSignature, signature.R...)
	concatenatedSignature = append(concatenatedSignature, signature.S...)
	concatenatedSignature = append(concatenatedSignature, signature.V)
	return
}
