// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type KnownAccounts map[common.Address]*Value

type Value struct {
	Single  *common.Hash
	Storage map[common.Hash]common.Hash
}

func SingleFromHex(hex string) *Value {
	return &Value{Single: common.HexToRefHash(hex)}
}

func FromMap(m map[string]string) *Value {
	res := map[common.Hash]common.Hash{}

	for k, v := range m {
		res[common.HexToHash(k)] = common.HexToHash(v)
	}

	return &Value{Storage: res}
}

func (v *Value) IsSingle() bool {
	return v != nil && v.Single != nil && !v.IsStorage()
}

func (v *Value) IsStorage() bool {
	return v != nil && v.Storage != nil
}

const EmptyValue = "{}"

func (v *Value) MarshalJSON() ([]byte, error) {
	if v.IsSingle() {
		return json.Marshal(v.Single)
	}

	if v.IsStorage() {
		return json.Marshal(v.Storage)
	}

	return []byte(EmptyValue), nil
}

const hashTypeName = "Hash"

func (v *Value) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var m map[string]json.RawMessage

	err := json.Unmarshal(data, &m)
	if err != nil {
		// single Hash value case
		v.Single = new(common.Hash)

		innerErr := json.Unmarshal(data, v.Single)
		if innerErr != nil {
			return fmt.Errorf("can't unmarshal to single value with error: %v value %q", innerErr, string(data))
		}

		return nil
	}

	res := make(map[common.Hash]common.Hash, len(m))

	for k, v := range m {
		// check k if it is a Hex value
		var kHash common.Hash

		err = hexutil.UnmarshalFixedText(hashTypeName, []byte(k), kHash[:])
		if err != nil {
			return fmt.Errorf("%w by key: %s with key %q and value %q", ErrKnownAccounts, err, k, string(v))
		}

		// check v if it is a Hex value
		var vHash common.Hash

		err = hexutil.UnmarshalFixedText("hashTypeName", bytes.Trim(v, "\""), vHash[:])
		if err != nil {
			return fmt.Errorf("%w by value: %s with key %q and value %q", ErrKnownAccounts, err, k, string(v))
		}

		res[kHash] = vHash
	}

	v.Storage = res

	return nil
}

func InsertKnownAccounts[T common.Hash | map[common.Hash]common.Hash](accounts KnownAccounts, k common.Address, v T) {
	switch typedV := any(v).(type) {
	case common.Hash:
		accounts[k] = &Value{Single: &typedV}
	case map[common.Hash]common.Hash:
		accounts[k] = &Value{Storage: typedV}
	}
}

type OptionsAA4337 struct {
	KnownAccounts  KnownAccounts `json:"knownAccounts"`
	BlockNumberMin *big.Int      `json:"blockNumberMin"`
	BlockNumberMax *big.Int      `json:"blockNumberMax"`
	TimestampMin   *uint64       `json:"timestampMin"`
	TimestampMax   *uint64       `json:"timestampMax"`
}

var ErrEmptyKnownAccounts = errors.New("knownAccounts cannot be nil")
var ErrKnownAccounts = errors.New("an incorrect list of knownAccounts")

func (ka KnownAccounts) ValidateLength() error {
	return nil

	// if ka == nil {
	// 	return ErrEmptyKnownAccounts
	// }

	// length := 0

	// for _, v := range ka {
	// 	// check if the value is hex string or an object
	// 	if v.IsSingle() {
	// 		length += 1
	// 	} else {
	// 		length += len(v.Storage)
	// 	}
	// }

	// if length >= 1000 {
	// 	return fmt.Errorf("number of slots/accounts in KnownAccounts %v exceeds the limit of 1000", length)
	// }

	// return nil
}
