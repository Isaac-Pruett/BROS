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

	"github.com/BooleanCat/option"
)

// A publisher that allows to send data.
//
// Publishers are automatically undeclared when dropped.
type Publisher struct {
	publisher *C.z_owned_publisher_t
}

// Options passed to [Publisher.Put] operation.
type PublisherPutOptions struct {
	Encoding    option.Option[Encoding]   // The encoding of the publication.
	Attachement option.Option[ZBytes]     // The attachment to attach to the publication.
	TimeStamp   option.Option[TimeStamp]  // The timestamp of the publication.
	SourceInfo  option.Option[SourceInfo] // Warning: This API has been marked as unstable. The source info for the publication.
}

func (opts *PublisherPutOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_publisher_put_options_t {
	var cOpts C.zc_cgo_publisher_put_options_t
	if opts.Attachement.IsSome() {
		opts.Attachement.Unwrap().toCData(pinner, &cOpts.attachment_data)
		cOpts.has_attachment = true
	}
	if opts.Encoding.IsSome() {
		opts.Encoding.Unwrap().toCData(pinner, &cOpts.encoding_data)
		cOpts.has_encoding = true
	}
	if opts.TimeStamp.IsSome() {
		cOpts.has_timestamp = true
		cOpts.timestamp = opts.TimeStamp.Unwrap().timestamp
	}
	if opts.SourceInfo.IsSome() {
		cOpts.has_source_info = true
		cOpts.source_info = opts.SourceInfo.Unwrap().sourceInfo
	}
	return cOpts
}

// Options passed to [Publisher.Delete] operation.
type PublisherDeleteOptions struct {
	TimeStamp option.Option[TimeStamp] // The timestamp of the publication.
}

func (opts *PublisherDeleteOptions) toCOpts(_ *runtime.Pinner) C.zc_cgo_publisher_delete_options_t {
	var cOpts C.zc_cgo_publisher_delete_options_t
	if opts.TimeStamp.IsSome() {
		cOpts.has_timestamp = true
		cOpts.timestamp = opts.TimeStamp.Unwrap().timestamp
	}
	return cOpts
}

// Undeclare and destroy the publisher.
func (publisher *Publisher) Undeclare() error {
	res := int8(C.z_undeclare_publisher(C.z_publisher_move(publisher.publisher)))
	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Destroy the publisher.
// This is equivalent to calling [Publisher.Undeclare] and discarding its return value.
func (publisher *Publisher) Drop() {
	C.z_publisher_drop(C.z_publisher_move(publisher.publisher))
}

// Publish message onto the publisher's key expression.
func (publisher *Publisher) Put(payload ZBytes, options *PublisherPutOptions) error {
	pinner := runtime.Pinner{}
	var cPayload C.zc_cgo_bytes_data_t
	payload.toCData(&pinner, &cPayload)
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_publisher_put(publisher.publisher, &cPayload, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_publisher_put(publisher.publisher, &cPayload, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Publish a `DELETE` message onto the publisher's key expression.
func (publisher *Publisher) Delete(options *PublisherDeleteOptions) error {
	pinner := runtime.Pinner{}
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_publisher_delete(publisher.publisher, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_publisher_delete(publisher.publisher, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Get the key expression of the publisher.
func (publisher *Publisher) KeyExpr() KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.z_publisher_keyexpr(C.z_publisher_loan(publisher.publisher)))
	return newKeyExprFromCDataPtr(&ke)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the publisher's entity global ID.
func (publisher *Publisher) Id() EntityGlobalId {
	return newEntityGlobalIdFromC(C.z_publisher_id(C.z_publisher_loan(publisher.publisher)))
}

// Options passed to publisher declaration.
type PublisherOptions struct {
	Encoding           option.Option[Encoding]          // Default encoding for messages put by this publisher.
	CongestionControl  option.Option[CongestionControl] // The congestion control to apply when routing messages from this publisher.
	Priority           option.Option[Priority]          // The priority of messages from this publisher.
	IsExpress          bool                             // If set to ``true``, the messages of this publisher will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	Reliability        option.Option[Reliability]       // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The publisher reliability.
	AllowedDestination option.Option[Locality]          // Restrict the subscribers which receive the publications from this publisher to the ones with compatible AllowedOrigin.
}

func (opts *PublisherOptions) toCOpts(pinner *runtime.Pinner) C.z_publisher_options_t {
	var cOpts C.z_publisher_options_t
	C.z_publisher_options_default(&cOpts)
	if opts.Encoding.IsSome() {
		cEncoding := opts.Encoding.Unwrap().toCPtr()
		pinner.Pin(cEncoding)
		cOpts.encoding = C.z_encoding_move(cEncoding)
	}
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	if opts.Reliability.IsSome() {
		cOpts.reliability = uint32(opts.Reliability.Unwrap())
	}
	if opts.AllowedDestination.IsSome() {
		cOpts.allowed_destination = uint32(opts.AllowedDestination.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	return cOpts
}

// Construct and declare a publisher for the given key expression.
//
// Data can be put and deleted with this publisher with the help of the
// [Publisher.Put] and [Publisher.Delete] functions.
// Publisher MUST be explicitly destroyed using [Publisher.Undeclare] or [Publisher.Drop] once it is no longer needed.
func (session *Session) DeclarePublisher(keyexpr KeyExpr, options *PublisherOptions) (Publisher, error) {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	var cPublisher C.z_owned_publisher_t

	if options == nil {
		res = int8(C.z_declare_publisher(C.z_session_loan(session.session), &cPublisher, C.z_view_keyexpr_loan(cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_publisher(C.z_session_loan(session.session), &cPublisher, C.z_view_keyexpr_loan(cKeyexpr), &cOpts))
	}
	pinner.Unpin()
	if res == 0 {
		return Publisher{publisher: &cPublisher}, nil
	}
	return Publisher{}, newZError(res)
}

// Options passed to [Session.Put] operation.
type PutOptions struct {
	Encoding           option.Option[Encoding]          // The encoding of the publication.
	Attachement        option.Option[ZBytes]            // The attachment to attach to the publication.
	TimeStamp          option.Option[TimeStamp]         // The timestamp of the publication.
	CongestionControl  option.Option[CongestionControl] // The congestion control to apply when routing the publication.
	Priority           option.Option[Priority]          // The priority of the publication.
	IsExpress          bool                             // If set to ``true``, the message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	Reliability        option.Option[Reliability]       // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The put operation reliability.
	AllowedDestination option.Option[Locality]          // Restrict the subscribers which receive the publication to the ones with compatible AllowedOrigin.
	SourceInfo         option.Option[SourceInfo]        // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The source info for the publication.
}

func (opts *PutOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_put_options_t {
	var cOpts C.zc_cgo_put_options_t
	if opts.Attachement.IsSome() {
		opts.Attachement.Unwrap().toCData(pinner, &cOpts.attachment_data)
		cOpts.has_attachment = true
	}
	if opts.Encoding.IsSome() {
		opts.Encoding.Unwrap().toCData(pinner, &cOpts.encoding_data)
		cOpts.has_encoding = true
	}
	if opts.TimeStamp.IsSome() {
		cOpts.has_timestamp = true
		cOpts.timestamp = opts.TimeStamp.Unwrap().timestamp
	}
	if opts.Priority.IsSome() {
		cOpts.priority = C.z_priority_t(opts.Priority.Unwrap())
	} else {
		cOpts.priority = C.z_priority_t(PriorityDefault)
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = C.z_congestion_control_t(opts.CongestionControl.Unwrap())
	} else {
		cOpts.congestion_control = C.z_congestion_control_t(CongestionControlDefault)
	}
	if opts.Reliability.IsSome() {
		cOpts.reliability = C.z_reliability_t(opts.Reliability.Unwrap())
	} else {
		cOpts.reliability = C.z_reliability_t(ReliabilityDefault)
	}
	if opts.AllowedDestination.IsSome() {
		cOpts.allowed_destination = C.z_locality_t(opts.AllowedDestination.Unwrap())
	} else {
		cOpts.allowed_destination = C.z_locality_t(LocalityDefault)
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	if opts.SourceInfo.IsSome() {
		cOpts.has_source_info = true
		cOpts.source_info = opts.SourceInfo.Unwrap().sourceInfo
	}
	return cOpts
}

// Options passed to [Session.Delete] operation.
type DeleteOptions struct {
	TimeStamp          option.Option[TimeStamp]         // The timestamp of the delete message.
	CongestionControl  option.Option[CongestionControl] // The congestion control to apply when routing the delete message.
	Priority           option.Option[Priority]          // The priority of the delete message.
	IsExpress          bool                             // If set to ``true``, the delete message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	Reliability        option.Option[Reliability]       // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The delete operation reliability.
	AllowedDestination option.Option[Locality]          // Restrict the subscribers which receive the delete message to the ones with compatible AllowedOrigin.
}

func (opts *DeleteOptions) toCOpts(_ *runtime.Pinner) C.zc_cgo_delete_options_t {
	var cOpts C.zc_cgo_delete_options_t
	if opts.TimeStamp.IsSome() {
		cOpts.has_timestamp = true
		cOpts.timestamp = opts.TimeStamp.Unwrap().timestamp
	}
	if opts.Priority.IsSome() {
		cOpts.priority = C.z_priority_t(opts.Priority.Unwrap())
	} else {
		cOpts.priority = C.z_priority_t(PriorityDefault)
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = C.z_congestion_control_t(opts.CongestionControl.Unwrap())
	} else {
		cOpts.congestion_control = C.z_congestion_control_t(CongestionControlDefault)
	}
	if opts.Reliability.IsSome() {
		cOpts.reliability = C.z_reliability_t(opts.Reliability.Unwrap())
	} else {
		cOpts.reliability = C.z_reliability_t(ReliabilityDefault)
	}
	if opts.AllowedDestination.IsSome() {
		cOpts.allowed_destination = C.z_locality_t(opts.AllowedDestination.Unwrap())
	} else {
		cOpts.allowed_destination = C.z_locality_t(LocalityDefault)
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	return cOpts
}

// Publish data on specified key expression.
func (session *Session) Put(keyExpr KeyExpr, payload ZBytes, options *PutOptions) error {
	pinner := runtime.Pinner{}
	var cPayload C.zc_cgo_bytes_data_t
	payload.toCData(&pinner, &cPayload)
	cKeyexpr := keyExpr.toCData(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_put(session.session, cKeyexpr, &cPayload, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_put(session.session, cKeyexpr, &cPayload, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Send request to delete data on specified key expression (used when working with [Zenoh storages]).
//
// [Zenoh storages]: https://zenoh.io/docs/manual/abstractions/#storage
func (session *Session) Delete(keyExpr KeyExpr, options *DeleteOptions) error {
	pinner := runtime.Pinner{}
	cKeyexpr := keyExpr.toCData(&pinner)
	res := int8(0)
	if options == nil {
		res = int8(C.zc_cgo_delete(session.session, cKeyexpr, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_delete(session.session, cKeyexpr, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}
