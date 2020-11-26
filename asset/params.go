// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package asset

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
	fingerprintTypeMerkleTree
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

func (r *RegistrationParams) SetFingerprintFromReader(reader io.Reader) error {
	h := sha3.New512()
	if _, err := io.Copy(h, reader); err != nil {
		return err
	}

	digest := h.Sum(nil)
	r.Fingerprint = fmt.Sprintf("%02d%s", fingerprintTypeSHA3512, hex.EncodeToString(digest[:]))
	return nil
}

func (r *RegistrationParams) SetFingerprintFromReaders(readers []io.Reader) error {

	if nil == readers || len(readers) == 0 {
		return ErrEmptyContent
	}

	length := len(readers)
	hashes := make([][]byte, length)

	for i := 0; i < length; i++ {
		h := sha3.New512()
		if _, err := io.Copy(h, readers[i]); err != nil {
			return err
		}

		digest := h.Sum(nil)
		hashes[i] = digest
	}

	return r.setFingerprintFromHashes(hashes)
}

func (r *RegistrationParams) SetFingerprintFromDataArray(contents [][]byte) error {
	if nil == contents || len(contents) == 0 {
		return ErrEmptyContent
	}

	length := len(contents)
	hashes := make([][]byte, length)

	for i := 0; i < length; i++ {
		hash := sha3.Sum512(contents[i])
		hashes[i] = hash[:]
	}

	return r.setFingerprintFromHashes(hashes)
}

func (r *RegistrationParams) setFingerprintFromHashes(hashes [][]byte) error {
	tree := buildMerkleTree(hashes, func(left, right []byte) []byte {
		data := append(left, right...)
		hash := sha3.Sum512(data)
		return hash[:]
	})

	if len(tree) == 0 {
		return errors.New("could not build merkle tree")
	}

	root := tree[len(tree)-1]
	r.Fingerprint = fmt.Sprintf("%02d%s", fingerprintTypeMerkleTree, base64.StdEncoding.EncodeToString(root))
	return nil
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
