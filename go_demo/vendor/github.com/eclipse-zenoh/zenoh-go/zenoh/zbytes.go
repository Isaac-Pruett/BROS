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
)

// A Zenoh data.
type ZBytes struct {
	data []byte
}

// Construct data from byte sequence.
func NewZBytes(data []byte) ZBytes {
	return ZBytes{data: data}
}

// Construct data from string.
func NewZBytesFromString(data string) ZBytes {
	return ZBytes{data: []byte(data)}
}

// Return the underlying byte sequence.
func (zbytes ZBytes) Bytes() []byte {
	return zbytes.data
}

// Convert data to string.
func (zbytes ZBytes) String() string {
	return string(zbytes.data[:])
}

// Get number of bytes that data contains.
func (zbytes ZBytes) Len() int {
	return len(zbytes.data)
}

func (zbytes ZBytes) toC() C.z_owned_bytes_t {
	var out C.z_owned_bytes_t
	if len(zbytes.data) > 0 {
		C.z_bytes_copy_from_buf(&out, (*C.uint8_t)(unsafe.Pointer(&zbytes.data[0])), C.size_t(len(zbytes.data)))
	} else {
		C.z_bytes_empty(&out)
	}
	return out
}

func (zbytes ZBytes) toCData(pinner *runtime.Pinner, out *C.zc_cgo_bytes_data_t) {
	if len(zbytes.data) > 0 {
		pinner.Pin(&zbytes.data[0])
		out.data = unsafe.Pointer(&zbytes.data[0])
		out.len = C.size_t(len(zbytes.data))
	}
}

//go:linkname zbytesToUnsafeCData
func zbytesToUnsafeCData(zbytes ZBytes, pinner *runtime.Pinner, out unsafe.Pointer) {
	zbytes.toCData(pinner, (*C.zc_cgo_bytes_data_t)(out))
}

func newZBytesFromC(cZbytes C.zc_cgo_bytes_data_t) ZBytes {
	var b ZBytes
	size := int(cZbytes.len)
	if size > 0 {
		b.data = C.GoBytes(cZbytes.data, C.int(size))
	} else if cZbytes.data != nil {
		size = (int)(C.z_bytes_len((*C.z_loaned_bytes_t)(cZbytes.data)))
		b.data = make([]uint8, size)
		if size > 0 {
			C.zc_cgo_bytes_read_all((*C.z_loaned_bytes_t)(cZbytes.data), (*C.uint8_t)(unsafe.Pointer(&b.data[0])))
		}
	}
	return b
}
