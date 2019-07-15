package asset

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
	"golang.org/x/crypto/sha3"
)

const (
	maxNameLength     = 64
	maxMetadataLength = 2048
)

const (
	fingerprintTypeUserDefined = iota
	fingerprintTypeSHA3512
)

var (
	ErrInvalidNameLength     = errors.New("property name not set or exceeds the maximum length (64 Unicode characters)")
	ErrInvalidMetadataLength = errors.New("property metadata exceeds the maximum length (1024 Unicode characters)")
	ErrEmptyContent          = errors.New("asset content is empty")
	ErrNullRegistrant        = errors.New("registrant is null")
)

type RegistrationParams struct {
	Name        string `json:"name" pack:"utf8"`
	Fingerprint string `json:"fingerprint" pack:"utf8"`
	Metadata    string `json:"metadata" pack:"utf8"`
	Registrant  string `json:"registrant" pack:"account"`
	Signature   string `json:"signature"`
}

func NewRegistrationParams(name string, metadata map[string]string) (*RegistrationParams, error) {
	keys := make([]string, 0, len(metadata))
	for k := range metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(metadata)*2)
	for _, key := range keys {
		val := metadata[key]
		if key == "" || val == "" {
			continue
		}
		parts = append(parts, key, val)
	}
	compactMetadata := strings.Join(parts, "\u0000")

	if utf8.RuneCountInString(name) > maxNameLength {
		return nil, ErrInvalidNameLength
	}

	if utf8.RuneCountInString(compactMetadata) > maxMetadataLength {
		return nil, ErrInvalidMetadataLength
	}

	return &RegistrationParams{
		Name:     name,
		Metadata: compactMetadata,
	}, nil
}

func (r *RegistrationParams) SetFingerprintFromFile(name string) error {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	return r.SetFingerprintFromData(content)
}

func (r *RegistrationParams) SetFingerprintFromData(content []byte) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	digest := sha3.Sum512(content)
	r.Fingerprint = fmt.Sprintf("%02d%s", fingerprintTypeSHA3512, hex.EncodeToString(digest[:]))
	return nil
}

func (r *RegistrationParams) SetFingerprint(fingerprint string) error {
	if len(fingerprint) == 0 {
		return ErrEmptyContent
	}
	r.Fingerprint = fmt.Sprintf("%02d%s", fingerprintTypeUserDefined, fingerprint)
	return nil
}

func (r *RegistrationParams) Sign(registrant account.Account) error {
	if registrant == nil {
		return ErrNullRegistrant
	}
	r.Registrant = registrant.AccountNumber()

	message, err := utils.Pack(r)
	if err != nil {
		return err
	}
	r.Signature = hex.EncodeToString(registrant.Sign(message))

	return nil
}
