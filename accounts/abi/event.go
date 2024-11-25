// Copyright 2016 The go-ethereum Authors
// (original work)
// Copyright 2024 The Erigon Authors
// (modifications)
// This file is part of Erigon.
//
// Erigon is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Erigon is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Erigon. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"fmt"
	"strings"

	libcommon "github.com/erigontech/erigon/erigon-lib/common"
	"github.com/erigontech/erigon/erigon-lib/crypto"
)

// Event is an event potentially triggered by the EVM's LOG mechanism. The Event
// holds type information (inputs) about the yielded output. Anonymous events
// don't get the signature canonical representation as the first LOG topic.
type Event struct {
	// Name is the event name used for internal representation. It's derived from
	// the raw name and a suffix will be added in the case of a event overload.
	//
	// e.g.
	// These are two events that have the same name:
	// * foo(int,int)
	// * foo(uint,uint)
	// The event name of the first one wll be resolved as foo while the second one
	// will be resolved as foo0.
	Name string
	// RawName is the raw event name parsed from ABI.
	RawName   string
	Anonymous bool
	Inputs    Arguments
	str       string
	// Sig contains the string signature according to the ABI spec.
	// e.g.	 event foo(uint32 a, int b) = "foo(uint32,int256)"
	// Please note that "int" is substitute for its canonical representation "int256"
	Sig string
	// ID returns the canonical representation of the event's signature used by the
	// abi definition to identify event names and types.
	ID libcommon.Hash
}

// NewEvent creates a new Event.
// It sanitizes the input arguments to remove unnamed arguments.
// It also precomputes the id, signature and string representation
// of the event.
func NewEvent(name, rawName string, anonymous bool, inputs Arguments) Event {
	// sanitize inputs to remove inputs without names
	// and precompute string and sig representation.
	names := make([]string, len(inputs))
	types := make([]string, len(inputs))
	for i, input := range inputs {
		if input.Name == "" {
			inputs[i] = Argument{
				Name:    fmt.Sprintf("arg%d", i),
				Indexed: input.Indexed,
				Type:    input.Type,
			}
		} else {
			inputs[i] = input
		}
		// string representation
		names[i] = fmt.Sprintf("%v %v", input.Type, inputs[i].Name)
		if input.Indexed {
			names[i] = fmt.Sprintf("%v indexed %v", input.Type, inputs[i].Name)
		}
		// sig representation
		types[i] = input.Type.String()
	}

	str := fmt.Sprintf("event %v(%v)", rawName, strings.Join(names, ", "))
	sig := fmt.Sprintf("%v(%v)", rawName, strings.Join(types, ","))
	id := libcommon.BytesToHash(crypto.Keccak256([]byte(sig)))

	return Event{
		Name:      name,
		RawName:   rawName,
		Anonymous: anonymous,
		Inputs:    inputs,
		str:       str,
		Sig:       sig,
		ID:        id,
	}
}

func (e Event) String() string {
	return e.str
}
