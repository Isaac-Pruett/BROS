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
import "github.com/BooleanCat/option"

// A Zenoh reply error - a combination of reply error payload and its encoding.
type ReplyError struct {
	payload  ZBytes
	encoding Encoding
}

// Return the error payload data.
func (reply_error *ReplyError) Payload() ZBytes {
	return reply_error.payload
}

// Return the encoding associated with the error data.
func (reply_error *ReplyError) Encoding() Encoding {
	return reply_error.encoding
}

type replyErr struct {
	value ReplyError
}

func (reply_error *replyErr) Ok() option.Option[Sample] {
	return option.None[Sample]()
}
func (reply_error *replyErr) Err() option.Option[ReplyError] {
	return option.Some(reply_error.value)
}
func (reply_error *replyErr) IsOk() bool {
	return false
}

type replyOk struct {
	value Sample
}

func (sample *replyOk) Ok() option.Option[Sample] {
	return option.Some(sample.value)
}
func (sample *replyOk) Err() option.Option[ReplyError] {
	return option.None[ReplyError]()
}
func (sample *replyOk) IsOk() bool {
	return true
}

// A Zenoh reply from a queryable.
type Reply interface {
	Ok() option.Option[Sample]      // Yield the contents of the reply by asserting it indicates a success.
	Err() option.Option[ReplyError] // Yield the contents of the reply by asserting it indicates an error.
	IsOk() bool                     // Return ``true`` if reply contains a valid response, ``false`` otherwise (in this case it contains an error value).
}

func newReplyFromC(cReplyData C.zc_cgo_reply_data_t) Reply {
	if cReplyData.is_ok {
		var s Sample
		s.payload = newZBytesFromC(cReplyData.payload)
		s.keyexpr = newKeyExprFromCDataPtr(&cReplyData.keyexpr)
		s.encoding = newEncodingFromC(cReplyData.encoding)
		s.kind = SampleKind(cReplyData.kind)
		s.reliability = Reliability(cReplyData.reliability)
		if cReplyData.timestamp != nil {
			s.timestamp = option.Some(TimeStamp{timestamp: *cReplyData.timestamp})
		}
		if cReplyData.attachment.len != 0 {
			s.attachment = option.Some(newZBytesFromC(cReplyData.attachment))
		}
		s.sourceInfo = newSourceInfoFromCPtr(cReplyData.source_info)
		return &replyOk{value: s}
	} else {
		var e ReplyError
		e.payload = newZBytesFromC(cReplyData.payload)
		e.encoding = newEncodingFromC(cReplyData.encoding)
		return &replyErr{value: e}
	}
}
