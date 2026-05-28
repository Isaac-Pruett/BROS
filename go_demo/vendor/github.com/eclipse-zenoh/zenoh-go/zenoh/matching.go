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
	"unsafe"

	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"
)

// A struct that indicates if there exist Subscribers matching the Publisher's key expression or Queryables matching Querier's key expression and target.
type MatchingStatus struct {
	Matching bool // ``True`` if there exist matching Zenoh entities, ``false`` otherwise.
}

//export zenohMatchingListenerCallback
func zenohMatchingListenerCallback(status *C.zc_cgo_const_matching_status, context unsafe.Pointer) {
	(*internal.ClosureContext[MatchingStatus])(context).Call(MatchingStatus{Matching: bool(status.matching)})
}

//export zenohMatchingListenerDrop
func zenohMatchingListenerDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[MatchingStatus])(context).Drop()
}

// A Zenoh matching listener.
//
// A listener that sends notifications when the [MatchingStatus] of a publisher or a querier changes.
type MatchingListener struct {
	listener *C.z_owned_matching_listener_t
	receiver <-chan MatchingStatus
}

func matchingListenerFromCPtrAndReceiver(listener *C.z_owned_matching_listener_t, receiver <-chan MatchingStatus) MatchingListener {
	return MatchingListener{listener: listener, receiver: receiver}
}

//go:linkname matchingListenerFromUnsafeCPtrAndReceiver
func matchingListenerFromUnsafeCPtrAndReceiver(listener unsafe.Pointer, receiver <-chan MatchingStatus) MatchingListener {
	return MatchingListener{listener: (*C.z_owned_matching_listener_t)(listener), receiver: receiver}
}

// Return matching listener's receiver if it was constructed with channel, nil otherwise.
func (listener *MatchingListener) Handler() <-chan MatchingStatus {
	return listener.receiver
}

// Destroy the matching listner.
// This is equivalent to calling [MatchingListener.Undeclare] and discarding its return value.
func (listener *MatchingListener) Drop() {
	C.z_matching_listener_drop(C.z_matching_listener_move(listener.listener))
}

// Undeclare and destroy the matching listener.
func (listener *MatchingListener) Undeclare() error {
	res := int8(C.z_undeclare_matching_listener(C.z_matching_listener_move(listener.listener)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Get publisher matching status - i.e. if there are any subscribers matching its key expression.
func (publisher *Publisher) GetMatchingStatus() (MatchingStatus, error) {
	var status C.z_matching_status_t
	res := int8(C.z_publisher_get_matching_status(C.z_publisher_loan(publisher.publisher), &status))

	if res == 0 {
		return MatchingStatus{Matching: bool(status.matching)}, nil
	}
	return MatchingStatus{}, newZError(res)
}

// Construct matching listener, registering a handler for notifying subscribers matching with a given publisher.
// Matching listener MUST be explicitly destroyed using [MatchingListener.Undeclare] or [MatchingListener.Drop] once it is no longer needed.
func (publisher *Publisher) DeclareMatchingListener(handler Handler[MatchingStatus]) (MatchingListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(closure))

	var cMatchingListener C.z_owned_matching_listener_t
	res := int8(C.z_publisher_declare_matching_listener(C.z_publisher_loan(publisher.publisher), &cMatchingListener, C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return MatchingListener{listener: &cMatchingListener, receiver: recv}, nil
	}
	return MatchingListener{}, newZError(res)
}

// Declare a matching listener, registering a callback for notifying subscribers matching with a given publisher.
// The callback will be run in the background until the corresponding publisher is dropped.
func (publisher *Publisher) DeclareBackgroundMatchingListener(closure Closure[MatchingStatus]) error {
	listenerClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	res := int8(C.z_publisher_declare_background_matching_listener(C.z_publisher_loan(publisher.publisher), C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Get querier matching status - i.e. if there are any queryables matching its key expression and target.
func (querier *Querier) GetMatchingStatus() (MatchingStatus, error) {
	var status C.z_matching_status_t
	res := int8(C.z_querier_get_matching_status(C.z_querier_loan(querier.querier), &status))

	if res == 0 {
		return MatchingStatus{Matching: bool(status.matching)}, nil
	}
	return MatchingStatus{}, newZError(res)
}

// Construct matching listener, registering a handler for notifying queryables matching with a given querier.
// Matching listener MUST be explicitly destroyed using [MatchingListener.Undeclare] or [MatchingListener.Drop] once it is no longer needed.
func (querier *Querier) DeclareMatchingListener(handler Handler[MatchingStatus]) (MatchingListener, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(closure))

	var cMatchingListener C.z_owned_matching_listener_t
	res := int8(C.z_querier_declare_matching_listener(C.z_querier_loan(querier.querier), &cMatchingListener, C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return MatchingListener{listener: &cMatchingListener, receiver: recv}, nil
	}
	return MatchingListener{}, newZError(res)
}

// Declare a matching listener, registering a callback for notifying queryables matching with a given querier.
// The callback will be run in the background until the corresponding publisher is dropped.
func (querier *Querier) DeclareBackgroundMatchingListener(closure Closure[MatchingStatus]) error {
	listenerClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_matching_status_t
	C.z_closure_matching_status(&cClosure, (*[0]byte)(C.zenohMatchingListenerCallback), (*[0]byte)(C.zenohMatchingListenerDrop), unsafe.Pointer(listenerClosure))

	res := int8(C.z_querier_declare_background_matching_listener(C.z_querier_loan(querier.querier), C.z_closure_matching_status_move(&cClosure)))

	if res == 0 {
		return nil
	}
	return newZError(res)
}
