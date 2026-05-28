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

	"github.com/eclipse-zenoh/zenoh-go/zenoh/internal"

	"github.com/BooleanCat/option"
)

//export zenohSubscriberCallbackData
func zenohSubscriberCallbackData(sample C.zc_cgo_sample_data_t, context unsafe.Pointer) {
	(*internal.ClosureContext[Sample])(context).Call(newSampleFromC(sample))
}

//export zenohSubscriberDrop
func zenohSubscriberDrop(context unsafe.Pointer) {
	(*internal.ClosureContext[Sample])(context).Drop()
}

// A Zenoh [subscriber].
// Receives data from publication on intersecting key expressions.
// Destroying the subscriber cancels the subscription.
//
// [subscriber]: https://zenoh.io/docs/manual/abstractions/#subscriber.
type Subscriber struct {
	subscriber *C.z_owned_subscriber_t
	receiver   <-chan Sample
}

func subscriberFromCPtrAndReceiver(subscriber *C.z_owned_subscriber_t, receiver <-chan Sample) Subscriber {
	return Subscriber{subscriber: subscriber, receiver: receiver}
}

//go:linkname subscriberFromUnsafeCPtrAndReceiver
func subscriberFromUnsafeCPtrAndReceiver(subscriber unsafe.Pointer, receiver <-chan Sample) Subscriber {
	return Subscriber{subscriber: (*C.z_owned_subscriber_t)(subscriber), receiver: receiver}
}

// Undeclare and destroy the subscriber.
func (subscriber *Subscriber) Undeclare() error {
	res := int8(C.z_undeclare_subscriber(C.z_subscriber_move(subscriber.subscriber)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Return Subscriber receiver if it was constructed with channel, nil otherwise.
func (subscriber *Subscriber) Handler() <-chan Sample {
	return subscriber.receiver
}

// Destroy the subscriber.
// This is equivalent to calling [Subscriber.Undeclare] and discarding its return value.
func (subscriber *Subscriber) Drop() {
	C.z_subscriber_drop(C.z_subscriber_move(subscriber.subscriber))
}

// Get the key expression of the subscriber.
func (subscriber *Subscriber) KeyExpr() KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.z_subscriber_keyexpr(C.z_subscriber_loan(subscriber.subscriber)))
	return newKeyExprFromCDataPtr(&ke)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the subscriber's entity global ID.
func (subscriber *Subscriber) Id() EntityGlobalId {
	return newEntityGlobalIdFromC(C.z_subscriber_id(C.z_subscriber_loan(subscriber.subscriber)))
}

// Options passed to subscriber declaration
type SubscriberOptions struct {
	AllowedOrigin option.Option[Locality] // Restrict the matching publications that will be received by this Subscriber to the ones that have compatible AllowedDestination.
}

func (opts *SubscriberOptions) toCOpts(_ *runtime.Pinner) C.z_subscriber_options_t {
	var cOpts C.z_subscriber_options_t
	C.z_subscriber_options_default(&cOpts)
	if opts.AllowedOrigin.IsSome() {
		cOpts.allowed_origin = uint32(opts.AllowedOrigin.Unwrap())
	}
	return cOpts
}

// Construct a subscriber for the given key expression.
// Subscriber MUST be explicitly destroyed using [Subscriber.Undeclare] or [Subscriber.Drop] once it is no longer needed.
func (session *Session) DeclareSubscriber(keyexpr KeyExpr, handler Handler[Sample], options *SubscriberOptions) (Subscriber, error) {
	callback, drop, recv := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(closure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	res := int8(0)
	var cSubscriber C.z_owned_subscriber_t
	if options == nil {
		res = int8(C.z_declare_subscriber(C.z_session_loan(session.session), &cSubscriber, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_subscriber(C.z_session_loan(session.session), &cSubscriber, C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Subscriber{subscriber: &cSubscriber, receiver: recv}, nil
	}
	return Subscriber{}, newZError(res)
}

// Construct and declare a background subscriber. Subscriber callback will be called to process the messages,
// until the corresponding session is closed or dropped.
func (session *Session) DeclareBackgroundSubscriber(keyexpr KeyExpr, closure Closure[Sample], options *SubscriberOptions) error {
	subClosure := internal.NewClosure(closure.Call, closure.Drop)
	var cClosure C.z_owned_closure_sample_t
	C.z_closure_sample(&cClosure, (*[0]byte)(C.zenohSubscriberCallback), (*[0]byte)(C.zenohSubscriberDrop), unsafe.Pointer(subClosure))
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.z_declare_background_subscriber(C.z_session_loan(session.session), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_background_subscriber(C.z_session_loan(session.session), C.z_view_keyexpr_loan(cKeyexpr), C.z_closure_sample_move(&cClosure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}
