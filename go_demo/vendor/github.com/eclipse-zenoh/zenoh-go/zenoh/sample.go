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

// The kind of Sample, can either be [SampleKindPut] or [SampleKindDelete].
type SampleKind int

const (
	SampleKindPut    SampleKind = C.Z_SAMPLE_KIND_PUT    // The Sample was issued by a ``Put`` operation.
	SampleKindDelete SampleKind = C.Z_SAMPLE_KIND_DELETE // The Sample was issued by a ``Delete`` operation.
)

// A Zenoh sample.
type Sample struct {
	keyexpr     KeyExpr
	payload     ZBytes
	kind        SampleKind
	encoding    Encoding
	timestamp   option.Option[TimeStamp]
	qos         qos
	attachment  option.Option[ZBytes]
	reliability Reliability
	sourceInfo  option.Option[SourceInfo]
}

// Return the key expression of the sample.
func (sample *Sample) KeyExpr() KeyExpr {
	return sample.keyexpr
}

// Return sample payload data.
func (sample *Sample) Payload() ZBytes {
	return sample.payload
}

// Return sample kind.
func (sample *Sample) Kind() SampleKind {
	return sample.kind
}

// Return the encoding associated with the sample data.
func (sample *Sample) Encoding() Encoding {
	return sample.encoding
}

// Return sample timestamp if there is any.
func (sample *Sample) TimeStamp() option.Option[TimeStamp] {
	return sample.timestamp
}

// Return sample attachment if there is any.
func (sample *Sample) Attachement() option.Option[ZBytes] {
	return sample.attachment
}

// Return sample qos priority value.
func (sample *Sample) Priority() Priority {
	return sample.qos.priority
}

// Return sample qos congestion contorl value.
func (sample *Sample) CongestionControl() CongestionControl {
	return sample.qos.congestionControl
}

// Return whether sample qos IsExpress flag was set or not.
func (sample *Sample) IsExpress() bool {
	return sample.qos.isExpress
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return sample qos reliability value.
func (sample *Sample) Reliability() Reliability {
	return sample.reliability
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Return the source info of the sample if present. Source info contains the global entity ID of the publisher
// and the sequence number of the sample.
func (sample *Sample) SourceInfo() option.Option[SourceInfo] {
	return sample.sourceInfo
}

func newSampleFromC(cSampleData C.zc_cgo_sample_data_t) Sample {
	var s Sample
	s.payload = newZBytesFromC(cSampleData.payload)
	s.keyexpr = newKeyExprFromCDataPtr(&cSampleData.keyexpr)
	s.encoding = newEncodingFromC(cSampleData.encoding)
	s.kind = SampleKind(cSampleData.kind)
	s.reliability = Reliability(cSampleData.reliability)
	if cSampleData.timestamp != nil {
		s.timestamp = option.Some(TimeStamp{timestamp: *cSampleData.timestamp})
	}
	if cSampleData.attachment.len != 0 {
		s.attachment = option.Some(newZBytesFromC(cSampleData.attachment))
	}
	s.sourceInfo = newSourceInfoFromCPtr(cSampleData.source_info)
	return s
}
