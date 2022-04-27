package boxstream

import (
	"fmt"
)

// NetworkKeyLength is implied to be constant due to an assumption that this
// key is used as an HMAC key when calling libsodium's crypto_auth during
// handshakes.
//
// See https://ssbc.github.io/scuttlebutt-protocol-guide/#handshake.
// See https://doc.libsodium.org/secret-key_cryptography/secret-key_authentication.
const NetworkKeyLength = 32

var defaultKey = [NetworkKeyLength]byte{0xd4, 0xa1, 0xcb, 0x88, 0xa6, 0x6f, 0x02, 0xf8, 0xdb, 0x63, 0x5c, 0xe2, 0x64, 0x41, 0xcc, 0x5d, 0xac, 0x1b, 0x08, 0x42, 0x0c, 0xea, 0xac, 0x23, 0x08, 0x39, 0xb7, 0x55, 0x84, 0x5a, 0x9f, 0xfb}

// NetworkKey is used for verifying that two peers are a part of the same
// Secure Scuttlebutt network in the initial stages of the handshake. Peers
// using two different network keys will not be able to establish a connection
// with each other. If you want to use the main Secure Scuttlebutt network then
// use NewDefaultNetworkKey. Setting a different network key using
// NewNetworkKey is mainly useful for test networks.
//
// See https://ssbc.github.io/scuttlebutt-protocol-guide/#handshake.
type NetworkKey struct {
	b []byte
}

// NewNetworkKey creates a network key from the provided slice of bytes. The
// slice must have a length of NetworkKeyLength.
func NewNetworkKey(b []byte) (NetworkKey, error) {
	if len(b) != NetworkKeyLength {
		return NetworkKey{}, fmt.Errorf("invalid transport key length, must be '%d'", NetworkKeyLength)
	}

	buf := make([]byte, NetworkKeyLength)
	copy(buf, b)
	return NetworkKey{buf}, nil
}

func MustNewNetworkKey(b []byte) NetworkKey {
	v, err := NewNetworkKey(b)
	if err != nil {
		panic(err)
	}
	return v
}

// NewDefaultNetworkKey creates a key initialized with an arbitrarily chosen
// value used in the default Secure Scuttlebutt network.
func NewDefaultNetworkKey() NetworkKey {
	return NetworkKey{defaultKey[:]}
}

func (k NetworkKey) Bytes() []byte {
	tmp := make([]byte, len(k.b))
	copy(tmp, k.b)
	return tmp
}

func (k NetworkKey) IsZero() bool {
	return len(k.b) == 0
}
