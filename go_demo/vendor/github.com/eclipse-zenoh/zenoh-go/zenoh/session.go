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
// typedef const z_id_t czid_t;
// void zenohZIdCallback(const z_id_t* id, void *context);
// void zenohZIdDrop(void *context);
import "C"
import (
	"unsafe"

	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

//export zenohZIdCallback
func zenohZIdCallback(id *C.czid_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Id])(context).Call(Id{id: *id})
}

//export zenohZIdDrop
func zenohZIdDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Id])(context).Drop()
}

// A Zenoh session.
type Session struct {
	session *C.z_owned_session_t
}

// Options passed to [Open] operation.
type SessionOptions struct {
}

func (options *SessionOptions) toCOpts() C.z_open_options_t {
	var opts C.z_open_options_t
	C.z_open_options_default(&opts)
	return opts
}

// Options passed to [Session.Close] operation.
type SessionCloseOptions struct {
}

func (*SessionCloseOptions) toCOpts() C.z_close_options_t {
	var opts C.z_close_options_t
	C.z_close_options_default(&opts)
	return opts
}

// Construct and open a new Zenoh session.
// Once work with session is finished, it MUST be explicitly destroyed by calling [Session.Drop].
func Open(config Config, options *SessionOptions) (Session, error) {
	var cSession C.z_owned_session_t
	c_config := config.toC()

	res := int8(0)
	if options == nil {
		res = int8(C.z_open(&cSession, C.z_config_move(&c_config), nil))
	} else {
		opts := options.toCOpts()
		res = int8(C.z_open(&cSession, C.z_config_move(&c_config), &opts))
	}

	if res == 0 {
		return Session{session: &cSession}, nil
	} else {
		return Session{}, newZError(res)
	}
}

// Close Zenoh session. This also calls drop functions for not yet dropped or undeclared Zenoh entites (subscribers, queryables, get queries).
// After this operation, all calls for network operations for entites declared on this session will return an error.
func (session *Session) Close(options *SessionCloseOptions) error {
	res := int8(0)
	if options == nil {
		res = int8(C.z_close(C.z_session_loan_mut(session.session), nil))
	} else {
		opts := options.toCOpts()
		res = int8(C.z_close(C.z_session_loan_mut(session.session), &opts))
	}
	if res == 0 {
		return nil
	} else {
		return newZError(res)
	}
}

// Checks if Zenoh session is closed.
func (session Session) IsClosed() bool {
	return bool(C.z_session_is_closed(C.z_session_loan(session.session)))
}

// Close and destroy the session. This MUST be called once all work with session is finished.
func (session *Session) Drop() {
	C.z_session_drop(C.z_session_move(session.session))
}

// Returns session's Zenoh ID.
func (session Session) ZId() Id {
	return Id{id: C.z_info_zid(C.z_session_loan(session.session))}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the session's entity global ID.
func (session Session) Id() EntityGlobalId {
	return newEntityGlobalIdFromC(C.z_session_id(C.z_session_loan(session.session)))
}

// Fetch Zenoh IDs of all connected peers.
func (session Session) PeersZId() ([]Id, error) {
	var ids []Id

	closure := internal.NewClosure(func(id Id) { ids = append(ids, id) }, nil)
	var cClosure C.z_owned_closure_zid_t
	C.z_closure_zid(&cClosure, (*[0]byte)(C.zenohZIdCallback), (*[0]byte)(C.zenohZIdDrop), unsafe.Pointer(closure))
	res := int8(C.z_info_peers_zid(C.z_session_loan(session.session), C.z_closure_zid_move(&cClosure)))
	if res != 0 {
		return []Id{}, newZError(res)
	}
	return ids, nil
}

// Fetch Zenoh IDs of all connected routers.
func (session Session) RoutersZId() ([]Id, error) {
	var ids []Id

	closure := internal.NewClosure(func(id Id) { ids = append(ids, id) }, nil)
	var cClosure C.z_owned_closure_zid_t
	C.z_closure_zid(&cClosure, (*[0]byte)(C.zenohZIdCallback), (*[0]byte)(C.zenohZIdDrop), unsafe.Pointer(closure))
	res := int8(C.z_info_routers_zid(C.z_session_loan(session.session), C.z_closure_zid_move(&cClosure)))
	if res != 0 {
		return []Id{}, newZError(res)
	}
	return ids, nil
}

func sessionGetInner(session *Session) unsafe.Pointer {
	return unsafe.Pointer(session.session)
}
