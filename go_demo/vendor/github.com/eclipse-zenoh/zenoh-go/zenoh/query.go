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

// A Zenoh query received by a queryable.
type Query struct {
	keyexpr        KeyExpr
	payload        option.Option[ZBytes]
	encoding       option.Option[Encoding]
	attachment     option.Option[ZBytes]
	parameters     string
	acceptsReplies ReplyKeyexpr
	sourceInfo     option.Option[SourceInfo]
	query          *C.z_owned_query_t
}

// Finalizes and destroys the query. This MUST be always called by user once all replies are provided.
func (query *Query) Drop() {
	C.zc_cgo_query_drop(query.query)
}

// Return the key expression of the query.
func (query *Query) KeyExpr() KeyExpr {
	return query.keyexpr
}

// Return query payload data if there is any.
func (query *Query) Payload() option.Option[ZBytes] {
	return query.payload
}

// Return the encoding associated with the query data, if there is any.
func (query *Query) Encoding() option.Option[Encoding] {
	return query.encoding
}

// Return query attachment if there is any.
func (query *Query) Attachement() option.Option[ZBytes] {
	return query.attachment
}

// Get query value selector parameters.
func (query *Query) Parameters() string {
	return query.parameters
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return the source info of the query if present. Source info contains the global entity ID of the querier
// and the sequence number of the query.
func (query *Query) SourceInfo() option.Option[SourceInfo] {
	return query.sourceInfo
}

func newQueryFromC(cQueryData C.zc_cgo_query_data_t) Query {
	var q Query
	q.keyexpr = newKeyExprFromCDataPtr(&cQueryData.keyexpr)
	q.parameters = C.GoStringN(cQueryData.params.str_ptr, C.int(cQueryData.params.len))
	if cQueryData.has_payload {
		q.payload = option.Some(newZBytesFromC(cQueryData.payload))
	}
	if cQueryData.has_attachment {
		q.attachment = option.Some(newZBytesFromC(cQueryData.attachment))
	}
	if cQueryData.has_encoding {
		q.encoding = option.Some(newEncodingFromC(cQueryData.encoding))
	}
	q.acceptsReplies = ReplyKeyexpr(cQueryData.accepts_replies)
	q.sourceInfo = newSourceInfoFromCPtr(cQueryData.source_info)
	q.query = &cQueryData.query
	return q
}

// Options passed to [Query.Reply] operation.
type QueryReplyOptions struct {
	Encoding          option.Option[Encoding]          // The encoding of the reply payload and/or attachment.
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // Deprecated: Congestion control setting is inherited from the query and cannot be overridden by the reply.
	Priority          option.Option[Priority]          // Deprecated: Priority setting is inherited from the query and cannot be overridden by the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	SourceInfo        option.Option[SourceInfo]        // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The source info for the reply.
}

func (opts *QueryReplyOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_query_reply_options_t {
	var cOpts C.zc_cgo_query_reply_options_t
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
	cOpts.is_express = C.bool(opts.IsExpress)
	if opts.SourceInfo.IsSome() {
		cOpts.has_source_info = true
		cOpts.source_info = opts.SourceInfo.Unwrap().sourceInfo
	}
	return cOpts
}

// Options passed to [Query.ReplyDel] operation.
type QueryReplyDelOptions struct {
	Attachement       option.Option[ZBytes]            // The attachment to attach to this reply.
	TimeStamp         option.Option[TimeStamp]         // The timestamp of the reply.
	CongestionControl option.Option[CongestionControl] // Deprecated: Congestion control setting is inherited from the query and cannot be overridden by the reply.
	Priority          option.Option[Priority]          // Deprecated: Priority setting is inherited from the query and cannot be overridden by the reply.
	IsExpress         bool                             // If set to ``true``, this reply message will not be batched. This usually has a positive impact on latency but negative impact on throughput.
	SourceInfo        option.Option[SourceInfo]        // Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release. The source info for the reply.
}

func (opts *QueryReplyDelOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_query_reply_del_options_t {
	var cOpts C.zc_cgo_query_reply_del_options_t
	if opts.Attachement.IsSome() {
		opts.Attachement.Unwrap().toCData(pinner, &cOpts.attachment_data)
		cOpts.has_attachment = true
	}
	if opts.TimeStamp.IsSome() {
		cOpts.has_timestamp = true
		cOpts.timestamp = opts.TimeStamp.Unwrap().timestamp
	}
	cOpts.is_express = C.bool(opts.IsExpress)
	if opts.SourceInfo.IsSome() {
		cOpts.has_source_info = true
		cOpts.source_info = opts.SourceInfo.Unwrap().sourceInfo
	}
	return cOpts
}

// Options passed to [Query.ReplyErr] operation.
type QueryReplyErrOptions struct {
	Encoding option.Option[Encoding] // The encoding of the reply payload.
}

func (opts *QueryReplyErrOptions) toCOpts(pinner *runtime.Pinner) C.zc_cgo_query_reply_err_options_t {
	var cOpts C.zc_cgo_query_reply_err_options_t
	if opts.Encoding.IsSome() {
		opts.Encoding.Unwrap().toCData(pinner, &cOpts.encoding_data)
		cOpts.has_encoding = true
	}
	return cOpts
}

// Send a reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) Reply(keyexpr KeyExpr, payload ZBytes, options *QueryReplyOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCData(&pinner)
	var cPayload C.zc_cgo_bytes_data_t
	payload.toCData(&pinner, &cPayload)
	if options == nil {
		res = int8(C.zc_cgo_query_reply(query.query, cKeyexpr, &cPayload, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply(query.query, cKeyexpr, &cPayload, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Send a delete reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) ReplyDel(keyexpr KeyExpr, options *QueryReplyDelOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	cKeyexpr := keyexpr.toCData(&pinner)
	if options == nil {
		res = int8(C.zc_cgo_query_reply_del(query.query, cKeyexpr, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply_del(query.query, cKeyexpr, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Send an error reply to the query.
//
// This function can be called multiple times to send multiple replies to a query. The reply
// will be considered complete when Drop is called.
func (query *Query) ReplyErr(payload ZBytes, options *QueryReplyErrOptions) error {
	res := int8(0)
	pinner := runtime.Pinner{}
	var cPayload C.zc_cgo_bytes_data_t
	payload.toCData(&pinner, &cPayload)
	if options == nil {
		res = int8(C.zc_cgo_query_reply_err(query.query, &cPayload, nil))
	} else {
		cOpts := options.toCOpts(&pinner)
		res = int8(C.zc_cgo_query_reply_err(query.query, &cPayload, &cOpts))
	}
	pinner.Unpin()

	if res == 0 {
		return nil
	}
	return newZError(res)
}

// Get the query `AcceptReplies` setting of this query, i.e. which replies are accepted by the query originator.
func (query *Query) AcceptsReplies() ReplyKeyexpr {
	return query.acceptsReplies
}
