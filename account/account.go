// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package account

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/sha3"
	"golang.org/x/text/language"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/encoding"
	"github.com/bitmark-inc/bitmarkd/util"
)

type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
)

const (
	ChecksumLength            = 4
	Base58AccountNumberLength = 37
)

const (
	pubkeyMask     = 0x01
	testnetMask    = 0x01 << 1
	algorithmShift = 4
)

const (
	seedHeaderLength   = 3
	seedPrefixLength   = 1
	seedCoreV1Length   = 32
	seedCoreV2Length   = 17
	seedChecksumLength = 4

	base58EncodedSeedV1Length     = 40
	base58EncodedseedCoreV2Length = 24

	recoveryPhraseV1Length   = 24
	recoveryPhraseV2Length   = 12
	recoveryPhraseV2CsLength = 13
)

var (
	seedHeader   = []byte{0x5a, 0xfe}
	seedHeaderV1 = append(seedHeader[:], []byte{0x01}...)
	seedHeaderV2 = append(seedHeader[:], []byte{0x02}...)

	// only for account v1
	seedNonce = [24]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	// only for account v1
	authSeedCount = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe7,
	}
	// only for account v1
	encrSeedCount = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
	}
)

var (
	ErrWrongNetwork          = errors.New("wrong network")
	ErrInvalidSeed           = errors.New("invalid seed")
	ErrInvalidRecoveryPhrase = errors.New("invalid recovery phrase")
	ErrInvalidChecksum       = errors.New("invalid checksum")
	ErrLangNotSupported      = errors.New("language not supported")
)

type Account interface {
	Version() Version
	Network() sdk.Network
	Seed() string
	RecoveryPhrase(language.Tag) ([]string, error)
	AccountNumber() string
	Bytes() []byte
	Sign(message []byte) (signature []byte)
}

func New() (Account, error) {
	// space for 128 bit random number, extended to 132 bits later
	seed := make([]byte, 16, seedCoreV2Length)

	n, err := rand.Read(seed)
	if err != nil {
		return nil, fmt.Errorf("only got: %d bytes expected: 16", n)
	}
	if n != 16 {
		return nil, fmt.Errorf("only got: %d bytes expected: 16", n)
	}

	// extend to 132 bits
	seed = append(seed, seed[15]&0xf0) // bits 7654xxxx  where x=zero

	// encode test/live flag
	mode := seed[0]&0x80 | seed[1]&0x40 | seed[2]&0x20 | seed[3]&0x10
	if sdk.GetNetwork() == sdk.Testnet {
		mode = mode ^ 0xf0
	}
	seed[15] = mode | seed[15]&0x0f

	return NewAccountV2(seed)
}

func FromSeed(seedBase58Encoded string) (Account, error) {
	s := encoding.FromBase58(seedBase58Encoded)

	if len(s) != base58EncodedSeedV1Length && len(s) != base58EncodedseedCoreV2Length {
		return nil, ErrInvalidSeed
	}

	data := s[:len(s)-seedChecksumLength]
	digest := sha3.Sum256(data)
	expectedChecksum := digest[:seedChecksumLength]
	actualChecksum := s[len(s)-seedChecksumLength:]
	if !bytes.Equal(expectedChecksum, actualChecksum) {
		return nil, ErrInvalidSeed
	}

	header := s[:seedHeaderLength]
	switch {
	case bytes.Equal(header, seedHeaderV1):
		// parse network
		prefix := s[seedHeaderLength : seedHeaderLength+seedPrefixLength]

		network := sdk.Livenet
		if prefix[0] == 0x01 {
			network = sdk.Testnet
		}

		if network != sdk.GetNetwork() {
			return nil, ErrWrongNetwork
		}

		seed := s[seedHeaderLength+seedPrefixLength : len(s)-seedChecksumLength]
		var core = new([32]byte)
		copy(core[:], seed)

		return NewAccountV1(core)
	case bytes.Equal(header, seedHeaderV2):
		// parse network
		var network sdk.Network
		core := s[seedHeaderLength : len(s)-seedChecksumLength]
		mode := core[0]&0x80 | core[1]&0x40 | core[2]&0x20 | core[3]&0x10
		switch mode {
		case core[15] & 0xF0:
			network = sdk.Livenet
		case core[15]&0xF0 ^ 0xF0:
			network = sdk.Testnet
		default:
			return nil, ErrInvalidSeed
		}

		if network != sdk.GetNetwork() {
			return nil, ErrWrongNetwork
		}

		checksumStart := len(s) - seedChecksumLength
		seed := s[seedHeaderLength:checksumStart]

		return NewAccountV2(seed)
	default:
		return nil, ErrInvalidSeed
	}
}

func FromRecoveryPhrase(words []string, lang language.Tag) (Account, error) {
	dict, err := getBIP39Dict(lang)
	if err != nil {
		return nil, err
	}

	switch len(words) {
	case recoveryPhraseV1Length:
		b, err := twentyFourWordsToBytes(words, dict)
		if err != nil {
			return nil, err
		}

		networkIndicator := b[0]
		var core = new([32]byte)
		copy(core[:], b[1:])

		var network sdk.Network
		switch networkIndicator {
		case 0x00:
			network = sdk.Livenet
		case 0x01:
			network = sdk.Testnet
		default:
			return nil, ErrInvalidRecoveryPhrase
		}

		if network != sdk.GetNetwork() {
			return nil, ErrWrongNetwork
		}

		return NewAccountV1(core)
	case recoveryPhraseV2Length:
		core, err := twelveWordsToBytes(words, dict)
		if err != nil {
			return nil, err
		}

		// parse network
		var network sdk.Network
		mode := core[0]&0x80 | core[1]&0x40 | core[2]&0x20 | core[3]&0x10
		switch mode {
		case core[15] & 0xF0:
			network = sdk.Livenet
		case core[15]&0xF0 ^ 0xF0:
			network = sdk.Testnet
		default:
			return nil, ErrInvalidSeed
		}

		if network != sdk.GetNetwork() {
			return nil, ErrWrongNetwork
		}

		return NewAccountV2(core)
	case recoveryPhraseV2CsLength:
		core, err := thirteenWordsToBytes(words, dict)
		if err != nil {
			return nil, err
		}

		// parse network
		var network sdk.Network
		mode := core[0]&0x80 | core[1]&0x40 | core[2]&0x20 | core[3]&0x10
		switch mode {
		case core[15] & 0xF0:
			network = sdk.Livenet
		case core[15]&0xF0 ^ 0xF0:
			network = sdk.Testnet
		default:
			return nil, ErrInvalidSeed
		}

		if network != sdk.GetNetwork() {
			return nil, ErrWrongNetwork
		}

		return NewAccountV2(core)
	default:
		return nil, ErrInvalidRecoveryPhrase
	}
}

type AccountV1 struct {
	network  sdk.Network
	seedCore *[32]byte
	AuthKey  AuthKey
	EncrKey  EncrKey
}

func NewAccountV1(seedCore *[seedCoreV1Length]byte) (*AccountV1, error) {
	authEntropy := secretbox.Seal([]byte{}, authSeedCount[:], &seedNonce, seedCore)
	authKey, err := NewAuthKey(authEntropy)
	if err != nil {
		return nil, err
	}

	encrEntropy := secretbox.Seal([]byte{}, encrSeedCount[:], &seedNonce, seedCore)
	encrKey, err := NewEncrKey(encrEntropy)
	if err != nil {
		return nil, err
	}

	return &AccountV1{sdk.GetNetwork(), seedCore, authKey, encrKey}, nil
}

func (acct *AccountV1) Network() sdk.Network {
	return acct.network
}

func (acct *AccountV1) Seed() string {
	var b bytes.Buffer
	b.Write(seedHeaderV1)

	seedPrefix := []byte{byte(0x00)}
	if acct.network == sdk.Testnet {
		seedPrefix = []byte{byte(0x01)}
	}
	b.Write(seedPrefix)

	b.Write(acct.seedCore[:])

	checksum := sha3.Sum256(b.Bytes())
	b.Write(checksum[:seedChecksumLength])

	return encoding.ToBase58(b.Bytes())
}

func (acct *AccountV1) RecoveryPhrase(lang language.Tag) ([]string, error) {
	dict, err := getBIP39Dict(lang)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	switch acct.Network() {
	case sdk.Livenet:
		buf.Write([]byte{00})
	case sdk.Testnet:
		buf.Write([]byte{01})
	}
	buf.Write(acct.seedCore[:])
	return bytesToTwentyFourWords(buf.Bytes(), dict)
}

func (acct *AccountV1) AccountNumber() string {
	buffer := acct.Bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:ChecksumLength]...)
	return encoding.ToBase58(buffer)
}

func (acct *AccountV1) Bytes() []byte {
	keyVariant := byte(acct.AuthKey.Algorithm()<<algorithmShift) | pubkeyMask
	if acct.network == sdk.Testnet {
		keyVariant |= testnetMask
	}
	return append([]byte{keyVariant}, acct.AuthKey.PublicKeyBytes()...)
}

func (acct *AccountV1) Sign(message []byte) []byte {
	return acct.AuthKey.Sign(message)
}

func (acct AccountV1) Version() Version {
	return V1
}

type AccountV2 struct {
	network  sdk.Network
	seedCore []byte
	AuthKey  AuthKey
	EncrKey  EncrKey
}

func NewAccountV2(seedCore []byte) (*AccountV2, error) {
	keys, err := seedCoreToKeys(seedCore, 2, 32)
	if err != nil {
		return nil, err
	}

	authKey, err := NewAuthKey(keys[0])
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(keys[1])
	if err != nil {
		return nil, err
	}

	return &AccountV2{
		network:  sdk.GetNetwork(),
		seedCore: seedCore,
		AuthKey:  authKey,
		EncrKey:  encrKey,
	}, nil
}

func seedCoreToKeys(seedCore []byte, keyCount int, keySize int) ([][]byte, error) {
	if len(seedCore) != seedCoreV2Length || seedCore[16]&0x0f != 0 {
		return nil, fmt.Errorf("invalid seed length")
	}

	if keyCount <= 0 {
		return nil, fmt.Errorf("invalid key count")
	}

	// add the seed 4 times to hash value
	hash := sha3.NewShake256()
	for i := 0; i < 4; i++ {
		n, err := hash.Write(seedCore)
		if err != nil {
			return nil, err
		}
		if n != seedCoreV2Length {
			return nil, fmt.Errorf("seed not successfully written: expected: %d bytes, actual: %d bytes", seedCoreV2Length, n)
		}
	}

	// generate count keys of size bytes
	keys := make([][]byte, keyCount)
	for i := 0; i < keyCount; i++ {
		k := make([]byte, keySize)
		n, err := hash.Read(k)
		if err != nil {
			return nil, err
		}
		if keySize != n {
			return nil, fmt.Errorf("key not successfully read: expected: %d bytes, actual: %d bytes", keySize, n)
		}
		keys[i] = k
	}
	return keys, nil
}

func (acct *AccountV2) Network() sdk.Network {
	return acct.network
}

func (acct *AccountV2) Seed() string {
	b := make([]byte, 0, seedHeaderLength+seedCoreV2Length+seedChecksumLength)

	b = append(b, seedHeaderV2...)
	b = append(b, acct.seedCore...)
	checksum := sha3.Sum256(b)
	b = append(b, checksum[:seedChecksumLength]...)
	b58Seed := util.ToBase58(b)

	return b58Seed
}

func (acct *AccountV2) RecoveryPhrase(lang language.Tag) ([]string, error) {
	dict, err := getBIP39Dict(lang)
	if err != nil {
		return nil, err
	}

	return bytesToThirteenWords(acct.seedCore, dict)
}

func (acct *AccountV2) AccountNumber() string {
	buffer := acct.Bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:ChecksumLength]...)
	return encoding.ToBase58(buffer)
}

func (acct *AccountV2) Bytes() []byte {
	keyVariant := byte(acct.AuthKey.Algorithm()<<algorithmShift) | pubkeyMask
	if acct.network == sdk.Testnet {
		keyVariant |= testnetMask
	}
	return append([]byte{keyVariant}, acct.AuthKey.PublicKeyBytes()...)
}

func (acct *AccountV2) Sign(message []byte) []byte {
	return acct.AuthKey.Sign(message)
}

func (acct AccountV2) Version() Version {
	return V2
}

func ValidateAccountNumber(accountNumber string) (err error) {
	_, err = extractAuthPublicKey(accountNumber)
	return
}

func Verify(accountNumber string, message, signature []byte) error {
	pubkey, err := extractAuthPublicKey(accountNumber)
	if err != nil {
		return err
	}

	if !ed25519.Verify(pubkey, message, signature) {
		return errors.New("invalid signature")
	}

	return nil
}

func extractAuthPublicKey(accountNumber string) (publicKey []byte, err error) {
	accountNumberBytes := encoding.FromBase58(accountNumber)
	if len(accountNumberBytes) == 0 {
		return nil, errors.New("invalid base58 string")
	}

	variantAndPubkey := accountNumberBytes[:len(accountNumberBytes)-ChecksumLength]
	computedChecksum := sha3.Sum256(variantAndPubkey)
	if !bytes.Equal(computedChecksum[:ChecksumLength], accountNumberBytes[len(accountNumberBytes)-ChecksumLength:]) {
		return nil, ErrInvalidChecksum
	}

	network := sdk.Livenet
	if accountNumberBytes[0]&testnetMask > 0 {
		network = sdk.Testnet
	}

	if network != sdk.GetNetwork() {
		return nil, ErrWrongNetwork
	}

	return variantAndPubkey[1:], nil
}
