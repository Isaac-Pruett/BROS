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
import "C"
import (
	"unsafe"

	"github.com/BooleanCat/option"
)

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Source information for a Zenoh message. Contains the global ID of the source entity and the sequence number.
type SourceInfo struct {
	sourceInfo C.z_source_info_t
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the global ID of the source entity.
func (si SourceInfo) Id() EntityGlobalId {
	return EntityGlobalId{id: C.z_source_info_id(&si.sourceInfo)}
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Returns the sequence number of the source message.
func (si SourceInfo) Sn() uint32 {
	return uint32(C.z_source_info_sn(&si.sourceInfo))
}

// Warning: This API has been marked as unstable: it works as advertised, but it may be changed in a future release.
//
// Creates a new SourceInfo from the given entity global ID and sequence number.
func NewSourceInfo(id EntityGlobalId, sn uint32) SourceInfo {
	return SourceInfo{sourceInfo: C.z_source_info_new(&id.id, C.uint32_t(sn))}
}

func newSourceInfoFromCPtr(si *C.z_source_info_t) option.Option[SourceInfo] {
	if si == nil {
		return option.None[SourceInfo]()
	}
	return option.Some(SourceInfo{sourceInfo: *si})
}

//go:linkname sourceInfoToUnsafeC
func sourceInfoToUnsafeC(si SourceInfo, cSourceInfo unsafe.Pointer) {
	*(*C.z_source_info_t)(cSourceInfo) = si.sourceInfo
}
