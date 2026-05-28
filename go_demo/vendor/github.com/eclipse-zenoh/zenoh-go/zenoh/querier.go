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

// Options passed to querier declaration.
type QuerierOptions struct {
	Target             option.Option[QueryTarget]        // The Queryables that should be target of the querier queries.
	Consolidataion     option.Option[QueryConsolidation] // The replies consolidation strategy to apply on replies to the querier queries.
	CongestionControl  option.Option[CongestionControl]  // The congestion control to apply when routing the query.
	Priority           option.Option[Priority]           // The priority of the querier queries.
	IsExpress          bool                              // If set to ``true``, the querier queries will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	TimeoutMs          uint64                            // The timeout for the querier queries in milliseconds. 0 means default query timeout from zenoh configuration.
	AllowedDestination option.Option[Locality]           // Restrict the queryables which receive the querier queries to the ones with compatible AllowedOrigin.
	AcceptReplies      option.Option[ReplyKeyexpr]       // The accepted replies for the querier queries.
}

func (opts *QuerierOptions) toCOpts(_pinner *runtime.Pinner) C.z_querier_options_t {
	var cOpts C.z_querier_options_t
	C.z_querier_options_default(&cOpts)
	if opts.Priority.IsSome() {
		cOpts.priority = uint32(C.z_priority_t(opts.Priority.Unwrap()))
	}
	if opts.CongestionControl.IsSome() {
		cOpts.congestion_control = uint32(opts.CongestionControl.Unwrap())
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	if opts.Target.IsSome() {
		cOpts.target = uint32(opts.Target.Unwrap())
	}
	if opts.Consolidataion.IsSome() {
		cOpts.consolidation.mode = int32(opts.Consolidataion.Unwrap().mode)
	}
	cOpts.timeout_ms = C.uint64_t(opts.TimeoutMs)
	if opts.AllowedDestination.IsSome() {
		cOpts.allowed_destination = uint32(opts.AllowedDestination.Unwrap())
	}
	if opts.AcceptReplies.IsSome() {
		cOpts.accept_replies = uint32(opts.AcceptReplies.Unwrap())
	}
	return cOpts
}

// A Zenoh querier.
//
// Sends queries to matching queryables.
type Querier struct {
	querier *C.z_owned_querier_t
}

// Options passed to [Querier.Get] operation.
type QuerierGetOptions struct {
	Payload           option.Option[ZBytes]            // An optional payload to attach to the query.
	Encoding          option.Option[Encoding]          // An optional encoding of the query payload and or attachment.
	Attachement       option.Option[ZBytes]            // The attachment to attach to the query.
	CancellationToken option.Option[CancellationToken] // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The cancellation token to interrupt the query.
	SourceInfo        option.Option[SourceInfo]        // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The source info for the query.
}

func (opts *QuerierGetOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_querier_get_options_t {
	var cOpts C.zc_cgo_querier_get_options_t

	if opts.Payload.IsSome() {
		opts.Payload.Unwrap().toCData(pinner, &cOpts.payload_data)
		cOpts.has_payload = true
	}
	if opts.Attachement.IsSome() {
		opts.Attachement.Unwrap().toCData(pinner, &cOpts.attachment_data)
		cOpts.has_attachment = true
	}
	if opts.Encoding.IsSome() {
		opts.Encoding.Unwrap().toCData(pinner, &cOpts.encoding_data)
		cOpts.has_encoding = true
	}
	if opts.CancellationToken.IsSome() {
		cOpts.cancellation_token = opts.CancellationToken.Unwrap().toC(pinner)
	}
	if opts.SourceInfo.IsSome() {
		cOpts.has_source_info = true
		cOpts.source_info = opts.SourceInfo.Unwrap().sourceInfo
	}

	return cOpts
}

// Get the key expression of the querier.
func (querier *Querier) KeyExpr() KeyExpr {
	ke := C.zc_cgo_keyexpr_get_data(C.z_querier_keyexpr(C.z_querier_loan(querier.querier)))
	return newKeyExprFromCDataPtr(&ke)
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the querier's entity global ID.
func (querier *Querier) Id() EntityGlobalId {
	return newEntityGlobalIdFromC(C.z_querier_id(C.z_querier_loan(querier.querier)))
}

// Construct a querier for the given key expression.
// Querier MUST be explicitly destroyed using [Querier.Drop] once it is no longer needed.
func (session *Session) DeclareQuerier(keyexpr KeyExpr, options *QuerierOptions) (Querier, error) {
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCPtr(&pinner)
	res := int8(0)
	var cQuerier C.z_owned_querier_t
	if options == nil {
		res = int8(C.z_declare_querier(C.z_session_loan(session.session), &cQuerier, C.z_view_keyexpr_loan(cKeyexpr), nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.z_declare_querier(C.z_session_loan(session.session), &cQuerier, C.z_view_keyexpr_loan(cKeyexpr), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return Querier{querier: &cQuerier}, nil
	}
	return Querier{}, newZError(res)
}

// Destroy the querier.
func (querier *Querier) Drop() {
	C.z_querier_drop(C.z_querier_move(querier.querier))
}

// Query data from the matching queryables in the system.
// Replies are provided through a callback function, if handler is a [Closure], through returned receiver if it is a [RingChannel] or a [FifoChannel].
func (querier *Querier) Get(parameters string, handler Handler[Reply], get_options *QuerierGetOptions) (<-chan Reply, error) {
	callback, drop, channel := handler.ToCbDropHandler()
	closure := internal.NewClosure(callback, drop)
	pinner := runtime.Pinner{}
	cParams := (*C.char)(nil)
	if len(parameters) != 0 {
		cParams = C.CString(parameters)
		defer C.free(unsafe.Pointer(cParams))
	}
	res := int8(0)
	if get_options == nil {
		res = int8(C.zc_cgo_querier_get(querier.querier, cParams, unsafe.Pointer(closure), nil))
	} else {
		cOpts := get_options.toCOpts(&pinner)
		res = int8(C.zc_cgo_querier_get(querier.querier, cParams, unsafe.Pointer(closure), &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return channel, nil
	}
	return nil, newZError(res)
}
