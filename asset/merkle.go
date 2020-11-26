// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC

package asset

func buildMerkleTree(txIds [][]byte, combineFunc func(left, right []byte) []byte) [][]byte {
	// compute length of ids + all tree levels including root
	idCount := len(txIds)

	totalLength := 1 // all ids + space for the final root
	for n := idCount; n > 1; n = (n + 1) / 2 {
		totalLength += n
	}

	// add initial ids
	tree := make([][]byte, totalLength)
	copy(tree[:], txIds)

	n := idCount
	j := 0
	for workLength := idCount; workLength > 1; workLength = (workLength + 1) / 2 {
		for i := 0; i < workLength; i += 2 {
			k := j + 1
			if i+1 == workLength {
				k = j // compensate for odd number
			}
			tree[n] = combineFunc(tree[j][:], tree[k][:])
			n += 1
			j = k + 1
		}
	}
	return tree
}
