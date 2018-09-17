package account

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/encoding"
	"golang.org/x/crypto/sha3"
)

const (
	pubkeyMask     = 0x01
	testnetMask    = 0x01 << 1
	algorithmShift = 4
	checksumLength = 4
)

var (
	seedHeader = []byte{0x5a, 0xfe, 0x01}
)

const (
	seedHeaderLength   = 3
	seedPrefixLength   = 1
	seedCoreLength     = 32
	seedChecksumLength = 4
	seedLength         = seedHeaderLength + seedPrefixLength + seedCoreLength + seedChecksumLength

	recoveryPhraseLength = 24
)

var (
	ErrInvalidNetwork = errors.New("invalid network")

	ErrSeedSizeMismatch     = errors.New("seed size mismatch")
	ErrSeedHeaderMismatch   = errors.New("seed header mismatch")
	ErrSeedChecksumMismatch = errors.New("seed checksum mismatch")

	ErrInvalidRecoveryPhrase = errors.New("invalid recovery phrase")
)

type Account struct {
	network sdk.Network
	core    *[32]byte
	AuthKey AuthKey
	EncrKey EncrKey
}

func New() (*Account, error) {
	var core [32]byte
	if _, err := io.ReadFull(rand.Reader, core[:]); err != nil {
		return nil, err
	}

	network := sdk.GetNetwork()

	return fromCore(network, &core)
}

func FromSeed(seed string) (*Account, error) {
	seedBytes := encoding.FromBase58(seed)

	if len(seedBytes) != seedLength {
		return nil, ErrSeedSizeMismatch
	}

	if !bytes.Equal(seedBytes[:seedHeaderLength], seedHeader) {
		return nil, ErrSeedHeaderMismatch
	}

	checksum := sha3.Sum256(seedBytes[:seedLength-seedChecksumLength])
	if !bytes.Equal(checksum[:seedChecksumLength], seedBytes[seedLength-seedChecksumLength:]) {
		return nil, ErrSeedChecksumMismatch
	}

	network := sdk.Livenet
	if seedBytes[seedHeaderLength : seedHeaderLength+seedPrefixLength][0] == 0x01 {
		network = sdk.Testnet
	}

	coreStart := seedHeaderLength + seedPrefixLength
	coreEnd := coreStart + seedCoreLength

	var core = new([32]byte)
	copy(core[:], seedBytes[coreStart:coreEnd])

	return fromCore(network, core)
}

func FromRecoveryPhrase(words []string) (*Account, error) {
	if len(words) != recoveryPhraseLength {
		return nil, ErrInvalidRecoveryPhrase
	}

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

	return fromCore(network, core)
}

func fromCore(network sdk.Network, core *[32]byte) (*Account, error) {
	if len(core) != 32 {
		return nil, ErrInvalidNetwork
	}
	authKey, err := NewAuthKey(core)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(core)
	if err != nil {
		return nil, err
	}

	return &Account{network, core, authKey, encrKey}, nil
}

func (acct *Account) Network() sdk.Network {
	return acct.network
}

func (acct *Account) Seed() string {
	var b bytes.Buffer
	b.Write(seedHeader)

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

func (acct *Account) RecoveryPhrase() []string {
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

func (acct *Account) AccountNumber() string {
	buffer := acct.Bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:checksumLength]...)
	return encoding.ToBase58(buffer)
}

func (acct *Account) Bytes() []byte {
	keyVariant := byte(acct.AuthKey.Algorithm()<<algorithmShift) | pubkeyMask
	if acct.network == sdk.Testnet {
		keyVariant |= testnetMask
	}
	return append([]byte{keyVariant}, acct.AuthKey.PublicKeyBytes()...)
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
