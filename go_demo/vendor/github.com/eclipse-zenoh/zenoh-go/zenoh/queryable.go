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
// #include "zenoh_cgo.h"
import "C"
import (
	"runtime"
	"unsafe"

	"github.com/BooleanCat/option"
	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

//export zenohQueryableCallbackData
func zenohQueryableCallbackData(query C.zc_cgo_query_data_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Query])(context).Call(newQueryFromC(query))
}

//export zenohQueryableDrop
func zenohQueryableDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Query])(context).Drop()
}

// A Zenoh [queryable]. Responds to queries sent via [Session.Get] with intersecting key expression.
//
// [queryable]: https://zenoh.io/docs/manual/abstractions/#queryable
type Queryable struct {
	queryable *C.z_owned_queryable_t
	receiver  <-chan Query
}

// Undeclare and destroy the queryable.
func (queryable *Queryable) Undeclare() error {
	res := int8(C.z_undeclare_queryable(C.z_queryable_move(queryable.queryable)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Return Queryable receiver if it was constructed with channel, nil otherwise.
func (queryable *Queryable) Handler() <-chan Query {
	return queryable.receiver
}

// Destroy the queryable.
// This is equivalent to calling [Queryable.Undeclare] and discarding its return value.
func (queryable *Queryable) Drop() {
	C.z_queryable_drop(C.z_queryable_move(queryable.queryable))
}

// Options passed to queryable declaration.
type QueryableOptions struct {
	Complete      bool                    // The completeness of the Queryable
	AllowedOrigin option.Option[Locality] // Restrict the matching requests that will be received by this Queryable to the ones that have compatible AllowedDestination.
}

func (opts *QueryableOptions) toCOpts(_ *runtime.Pinner) C.z_queryable_options_t {
	var cOpts C.z_queryable_options_t
	C.z_queryable_options_default(&cOpts)
	cOpts.complete = C.bool(opts.Complete)
	if opts.AllowedOrigin.IsSome() {
		cOpts.allowed_origin = uint32(opts.AllowedOrigin.Unwrap())
	}
	return cOpts
}

// Construct a queryable for the given key expression.
// Queryable MUST be explicitly destroyed using [Queryable.Undeclare] or [Queryable.Drop] once it is no longer needed.
func (session *Session) DeclareQueryable(keyexpr KeyExpr, handler Handler[Query], options *QueryableOptions) (Queryable, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_query_t
	C.z_closure_query(&cClosure, (*[0]byte)(C.zenohQueryableCallback), (*[0]byte)(C.zenohQueryableDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	res := int8(0)
	var cQueryable C.z_owned_queryable_t
	if options == nil {
		res = int8(C.z_declare_queryable(C.z_session_loan(session.session), &cQueryable, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_query_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_queryable(C.z_session_loan(session.session), &cQueryable, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_query_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Queryable{queryable: &cQueryable, receiver: channel}, nil
	}
	return Queryable{}, newZError(res)
}

// Declare a background queryable for a given keyexpr. The queryable callback will be be called
// to proccess incoming queries until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundQueryable(keyexpr KeyExpr, closure Closure[Query], options *QueryableOptions) error {
	qClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_query_t
	C.z_closure_query(&cClosure, (*[0]byte)(C.zenohQueryableCallback), (*[0]byte)(C.zenohQueryableDrop), unsafe.Pointer(qClosure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_declare_background_queryable(C.z_session_loan(session.session), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_query_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_background_queryable(C.z_session_loan(session.session), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_query_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Get the key expression of the queryable.
func (queryable *Queryable) KeyExpr() KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.z_queryable_keyexpr(C.z_queryable_loan(queryable.queryable)))
	return newKeyExprFromCDataPtr(&ke)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the queryable's entity global ID.
func (queryable *Queryable) Id() EntityGlobalId {
	return newEntityGlobalIdFromC(C.z_queryable_id(C.z_queryable_loan(queryable.queryable)))
}
