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
import "unsafe"

// Initializes the zenoh runtime logger, using rust environment settings or the provided fallback level.
// E.g.: `RUST_LOG=info` will enable logging at info level. Similarly, you can set the variable to `error` or `debug`.
//
// Note that if the environment variable is not set, then fallback [filter] will be used instead.
//
// [filter] https://docs.rs/env_logger/latest/env_logger/index.html.
func InitLoggerFromEnvOr(filter string) {
	cFilter := C.CString(filter)
	defer C.free(unsafe.Pointer(cFilter))
	C.zc_init_log_from_env_or(cFilter)
}
