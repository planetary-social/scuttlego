package boxstream

import (
	"fmt"
)

// networkKeyLength is implied to be constant due to an assumption that this key is used as an hmac key when calling
// libsodium's crypto_auth during handshakes.
// https://doc.libsodium.org/secret-key_cryptography/secret-key_authentication
// https://ssbc.github.io/scuttlebutt-protocol-guide/#handshake
const networkKeyLength = 32

var defaultKey = [networkKeyLength]byte{0xd4, 0xa1, 0xcb, 0x88, 0xa6, 0x6f, 0x02, 0xf8, 0xdb, 0x63, 0x5c, 0xe2, 0x64, 0x41, 0xcc, 0x5d, 0xac, 0x1b, 0x08, 0x42, 0x0c, 0xea, 0xac, 0x23, 0x08, 0x39, 0xb7, 0x55, 0x84, 0x5a, 0x9f, 0xfb}

type NetworkKey struct {
	b []byte
}

func NewKey(b []byte) (NetworkKey, error) {
	if len(b) != networkKeyLength {
		return NetworkKey{}, fmt.Errorf("invalid network key length, must be '%d'", networkKeyLength)
	}

	var buf []byte
	copy(buf, b)
	return NetworkKey{buf}, nil
}

func NewDefaultNetworkKey() NetworkKey {
	return NetworkKey{defaultKey[:]}
}

func (k NetworkKey) Bytes() []byte {
	tmp := make([]byte, len(k.b))
	copy(tmp, k.b)
	return tmp
}
