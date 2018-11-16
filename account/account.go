package account

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account/bip39"
	"github.com/bitmark-inc/bitmark-sdk-go/encoding"
	"github.com/bitmark-inc/bitmarkd/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/sha3"
)

const (
	pubkeyMask     = 0x01
	testnetMask    = 0x01 << 1
	algorithmShift = 4
	checksumLength = 4
)

var (
	seedHeader   = []byte{0x5a, 0xfe}
	seedHeaderV1 = append(seedHeader[:], []byte{0x01}...)
	seedHeaderV2 = append(seedHeader[:], []byte{0x02}...)
)

const (
	seedHeaderLength   = 3
	seedPrefixLength   = 1
	seedV1Length       = 32
	seedV2Length       = 17
	seedChecksumLength = 4

	recoveryPhraseV1Length = 24
	recoveryPhraseV2Length = 12
)

var (
	ErrInvalidNetwork = errors.New("invalid network")

	ErrInvalidSeedLength   = errors.New("invalid seed length")
	ErrInvalidSeedHeader   = errors.New("invalid seed header")
	ErrInvalidSeedChecksum = errors.New("invalid seed checksum")

	ErrInvalidRecoveryPhrase = errors.New("invalid recovery phrase")
)

type Account interface {
	Network() sdk.Network
	Seed() string
	RecoveryPhrase(string) []string
	AccountNumber() string
	Bytes() []byte
	Sign(message []byte) (signature []byte)
}

func New() (Account, error) {
	// space for 128 bit random number, extended to 132 bits later
	seed := make([]byte, 16, seedV2Length)

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

	// TODO
	if len(s) != 40 && len(s) != 24 {
		return nil, ErrInvalidSeedLength
	}

	data := s[:len(s)-seedChecksumLength]
	digest := sha3.Sum256(data)
	expectedChecksum := digest[:seedChecksumLength]
	actualChecksum := s[len(s)-seedChecksumLength:]
	if !bytes.Equal(expectedChecksum, actualChecksum) {
		return nil, ErrInvalidSeedChecksum
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
			return nil, fmt.Errorf("tried to recover %s account but the config is set to %s", network, sdk.GetNetwork())
		}

		seed := s[seedHeaderLength+seedPrefixLength : len(s)-seedChecksumLength]
		var core = new([32]byte)
		copy(core[:], seed)

		return NewAccountV1(core)
	case bytes.Equal(header, seedHeaderV2):
		checksumStart := len(s) - seedChecksumLength
		seed := s[seedHeaderLength:checksumStart]

		return NewAccountV2(seed)
	default:
		return nil, ErrInvalidSeedLength
	}
}

func FromRecoveryPhrase(words []string, lang string) (Account, error) {
	switch len(words) {
	case recoveryPhraseV1Length:
		b, err := phraseToBytes(words)
		if err != nil {
			return nil, err
		}

		networkIndicator := b[0]
		var core = new([32]byte)
		copy(core[:], b[1:])

		network, err := parseNetwork(networkIndicator)
		if err != nil {
			return nil, err
		}

		if network != sdk.GetNetwork() {
			return nil, fmt.Errorf("tried to recover %s account but the config is set to %s", network, sdk.GetNetwork())
		}

		return NewAccountV1(core)
	case recoveryPhraseV2Length:
		dict := getBIP39Dict(lang)

		seed := make([]byte, 0, 17)

		remainder := 0
		bits := 0
		for _, word := range words {
			n := -1
		loop:
			for i, bip := range dict {
				if word == bip {
					n = i
					break loop
				}
			}
			if n < 0 {
				return nil, fmt.Errorf("invalid word: %q", word)
			}
			remainder = remainder<<11 + n
			for bits += 11; bits >= 8; bits -= 8 {
				a := 0xff & (remainder >> uint(bits-8))
				seed = append(seed, byte(a))
			}
			remainder &= masks[bits]
		}

		// check that the whole 16 bytes are converted and the final nibble remains to be packed
		if 4 != bits || 16 != len(seed) {
			return nil, fmt.Errorf("only converted: %d bytes expected: 16.5", len(seed))
		}

		// justify final 4 bits to high nibble, low nibble is zero
		seed = append(seed, byte(remainder<<4))

		return NewAccountV2(seed)
	default:
		return nil, ErrInvalidRecoveryPhrase
	}
}

type AccountV1 struct {
	network sdk.Network
	core    *[32]byte
	AuthKey AuthKey
	EncrKey EncrKey
}

func NewAccountV1(seed *[32]byte) (*AccountV1, error) {
	if len(seed) != 32 {
		return nil, ErrInvalidNetwork
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	return &AccountV1{sdk.GetNetwork(), seed, authKey, encrKey}, nil
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

	b.Write(acct.core[:])

	checksum := sha3.Sum256(b.Bytes())
	b.Write(checksum[:seedChecksumLength])

	return encoding.ToBase58(b.Bytes())
}

func (acct *AccountV1) RecoveryPhrase(lang string) []string {
	buf := new(bytes.Buffer)
	switch acct.Network() {
	case sdk.Livenet:
		buf.Write([]byte{00})
	case sdk.Testnet:
		buf.Write([]byte{01})
	}
	buf.Write(acct.core[:])
	phrase, _ := bytesToPhrase(buf.Bytes())
	return phrase
}

func (acct *AccountV1) AccountNumber() string {
	buffer := acct.Bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:checksumLength]...)
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

type AccountV2 struct {
	network sdk.Network
	seed    []byte
	AuthKey AuthKey
	EncrKey EncrKey
}

func NewAccountV2(seed []byte) (*AccountV2, error) {
	keys, err := seedToKeys(seed, 2, 32)
	if err != nil {
		return nil, err
	}

	_, authPvtKey, err := ed25519.GenerateKey(bytes.NewBuffer(keys[0]))
	if err != nil {
		return nil, err
	}
	authKey := ED25519AuthKey{authPvtKey}

	encrPubKey, encrPvtKey, err := box.GenerateKey(bytes.NewBuffer(keys[1]))
	if err != nil {
		return nil, err
	}
	encrKey := CURVE25519EncrKey{encrPubKey, encrPvtKey}

	return &AccountV2{
		network: sdk.GetNetwork(),
		seed:    seed,
		AuthKey: authKey,
		EncrKey: encrKey,
	}, nil
}

func seedToKeys(seed []byte, keyCount int, keySize int) ([][]byte, error) {
	if len(seed) != seedV2Length || seed[16]&0x0f != 0 {
		return nil, fmt.Errorf("invalid seed length: expected: %d bytes, actual: %d bytes", seedV2Length, len(seed))
	}

	if keyCount <= 0 {
		return nil, fmt.Errorf("invalid key count")
	}

	// add the seed 4 times to hash value
	hash := sha3.NewShake256()
	for i := 0; i < 4; i += 1 {
		n, err := hash.Write(seed)
		if err != nil {
			return nil, err
		}
		if n != seedV2Length {
			return nil, fmt.Errorf("seed not successfully written: expected: %d bytes, actual: %d bytes", seedV2Length, n)
		}
	}

	// generate count keys of size bytes
	keys := make([][]byte, keyCount)
	for i := 0; i < keyCount; i += 1 {
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
	b := make([]byte, 0, seedHeaderLength+seedV2Length+seedChecksumLength)

	b = append(b, seedHeaderV2...)
	b = append(b, acct.seed...)
	checksum := sha3.Sum256(b)
	b = append(b, checksum[:seedChecksumLength]...)
	b58Seed := util.ToBase58(b)

	return b58Seed
}

func (acct *AccountV2) RecoveryPhrase(lang string) []string {
	phrase := make([]string, 0, 12)
	accumulator := 0
	bits := 0
	n := 0
	for i := 0; i < len(acct.seed); i += 1 {
		accumulator = accumulator<<8 + int(acct.seed[i])
		bits += 8
		if bits >= 11 {
			bits -= 11 // [ 11 bits] [offset bits]

			n += 1
			index := accumulator >> uint(bits)
			accumulator &= masks[bits]

			var word string
			switch lang {
			case "en":
				word = bip39.English[index]
			case "zh-TW":
				word = bip39.TraditionalChinese[index]
			default:
				word = bip39.English[index]
			}

			phrase = append(phrase, word)
		}
	}
	return phrase
}

func (acct *AccountV2) AccountNumber() string {
	buffer := acct.Bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:checksumLength]...)
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

func parseNetwork(b byte) (sdk.Network, error) {
	switch b {
	case 0x00:
		return sdk.Livenet, nil
	case 0x01:
		return sdk.Testnet, nil
	default:
		return "", ErrInvalidNetwork
	}
}

func ParseAccountNumber(number string) (sdk.Network, []byte, error) {
	accountNumberBytes := encoding.FromBase58(number)

	acct := accountNumberBytes[:len(accountNumberBytes)-checksumLength]
	computedChecksum := sha3.Sum256(acct)
	if !bytes.Equal(computedChecksum[:checksumLength], accountNumberBytes[len(accountNumberBytes)-checksumLength:]) {
		return "", nil, errors.New("invalid account number")
	}

	network := sdk.Livenet
	if accountNumberBytes[0]&testnetMask > 0 {
		network = sdk.Testnet
	}

	return network, acct, nil
}

func getBIP39Dict(lang string) []string {
	switch lang {
	case "en":
		return bip39.English
	case "zh-TW":
		return bip39.TraditionalChinese
	default:
		return bip39.English
	}
}
