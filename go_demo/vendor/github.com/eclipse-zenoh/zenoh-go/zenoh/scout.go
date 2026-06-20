//
// Copyright (c) 2025 ZettaScale Technology
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
//
// Contributors:
//   ZettaScale Zenoh Team, <zenoh@zettascale.tech>
//

package zenoh

// #include "zenoh.h"
// void zenohScoutCallback(struct z_loaned_hello_t *sample, void *context);
// void zenohScoutDrop(void *context);
// static const z_what_t CGO_DEFAULT_SCOUTING_WHAT = Z_WHAT_ROUTER_PEER;
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

//export zenohScoutCallback
func zenohScoutCallback(hello *C.z_loaned_hello_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Hello])(context).Call(newHelloFromC(hello))
}

//export zenohScoutDrop
func zenohScoutDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Hello])(context).Drop()
}

// Enum indicating type of Zenoh entity.
type WhatAmI uint32

const (
	WhatAmIRouter WhatAmI = C.Z_WHATAMI_ROUTER // Entity type is Router.
	WhatAmIPeer   WhatAmI = C.Z_WHATAMI_PEER   // Entity type is Peer.
	WhatAmIClient WhatAmI = C.Z_WHATAMI_CLIENT // Entity type is Client.
)

func (whatami WhatAmI) String() string {
	switch whatami {
	case WhatAmIRouter:
		return "Router"
	case WhatAmIClient:
		return "Client"
	case WhatAmIPeer:
		return "Peer"
	default:
		return "unknown"
	}
}

// Flag indicating type of Zenoh entities to scout for.
type What uint32

const (
	WhatRouter           What = C.Z_WHAT_ROUTER             // Scout for Routers.
	WhatPeer             What = C.Z_WHAT_PEER               // Scout for Peers.
	WhatClient           What = C.Z_WHAT_CLIENT             // Scout for Clients.
	WhatRouterPeer       What = C.Z_WHAT_ROUTER_PEER        // Scout for Routers and Peers.
	WhatRouterClient     What = C.Z_WHAT_ROUTER_CLIENT      // Scout for Routers and Clients.
	WhatPeerClient       What = C.Z_WHAT_PEER_CLIENT        // Scout for Peers and Clients.
	WhatRouterPeerClient What = C.Z_WHAT_ROUTER_PEER_CLIENT // Scout for Routers, Peers and Clients.
	WhatDefault          What = C.CGO_DEFAULT_SCOUTING_WHAT // Use default scouting mask.
)

// A Zenoh hello message replied by a Zenoh entity to a scout message sent with [Scout].
type Hello struct {
	whatAmI  WhatAmI
	locators []string
	id       Id
}

func newHelloFromC(cHello *C.z_loaned_hello_t) Hello {
	var cLocators C.z_owned_string_array_t
	C.z_hello_locators(cHello, &cLocators)
	loanedCLocators := C.z_string_array_loan(&cLocators)
	locators := make([]string, int(C.z_string_array_len(loanedCLocators)))

	for i := 0; i < len(locators); i++ {
		s := C.z_string_array_get(loanedCLocators, C.size_t(i))
		locators[i] = C.GoStringN(C.z_string_data(s), C.int(C.z_string_len(s)))
	}

	return Hello{whatAmI: WhatAmI(C.z_hello_whatami(cHello)), locators: locators, id: Id{id: C.z_hello_zid(cHello)}}
}

// Get array of locators of Zenoh entity that transmitted hello message.
func (hello Hello) Locators() []string {
	return hello.locators
}

// Get ID of Zenoh entity that transmitted hello message.
func (hello Hello) ZId() Id {
	return hello.id
}

// Get type of Zenoh entity that transmitted hello message.
func (hello Hello) WhatAmI() WhatAmI {
	return hello.whatAmI
}

func (hello Hello) String() string {
	return fmt.Sprintf("Hello { pid: %s, whatami: %s, locators: %s}", hello.id, hello.whatAmI, hello.locators)
}

// Options passed to [Scout] operation.
type ScoutOptions struct {
	TimeoutMs uint64 // The maximum duration in ms the scouting can take. 0 corresponds to default scouting timeout value.
	What      What   // Type of entities to scout for. 0 corresponds to default scouting entities type.
}

func (opts *ScoutOptions) toCOpts() C.z_scout_options_t {
	var cOpts C.z_scout_options_t
	C.z_scout_options_default(&cOpts)
	if opts.TimeoutMs != 0 {
		cOpts.timeout_ms = C.uint64_t(opts.TimeoutMs)
	}
	if opts.What != 0 {
		cOpts.what = uint32(opts.What)
	}
	return cOpts
}

// Scout for routers and/or peers.
// This function will block if handler returns a nil channel part and will run in the background otherwise.
func Scout(config Config, handler Handler[Hello], options *ScoutOptions) (<-chan Hello, error) {
	var callback, drop, channel = handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_hello_t
	C.z_closure_hello(&cClosure, (*[0]byte)(C.zenohScoutCallback), (*[0]byte)(C.zenohScoutDrop), unsafe.Pointer(closure))
	cConfig := config.toC()
	res := int8(0)
	if options == nil {
		if channel == nil {
			res = int8(C.z_scout(C.z_config_move(&cConfig), C.z_closure_hello_move(&cClosure), nil))
		} else {
			go C.z_scout(C.z_config_move(&cConfig), C.z_closure_hello_move(&cClosure), nil)
		}
	} else {
		cOpts := options.toCOpts()
		if channel == nil {
			res = int8(C.z_scout(C.z_config_move(&cConfig), C.z_closure_hello_move(&cClosure), &cOpts))
		} else {
			go C.z_scout(C.z_config_move(&cConfig), C.z_closure_hello_move(&cClosure), &cOpts)
		}
	}

	if res == 0 {
		return channel, nil
	}
	return nil, newZError(res)
}
