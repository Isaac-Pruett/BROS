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

package internal

// #include <stdint.h>
import "C"
import (
	"runtime"
	"runtime/cgo"
)

type ClosureContext[T any] struct {
	onCall C.uintptr_t
	onDrop C.uintptr_t
	pinner C.uintptr_t
}

func (context *ClosureContext[T]) Call(value T) {
	cgo.Handle(context.onCall).Value().(func(T))(value)
}

func (context *ClosureContext[T]) Drop() {
	if C.uintptr_t(context.onDrop) != 0 {
		cgo.Handle(context.onDrop).Value().(func())()
		cgo.Handle(context.onDrop).Delete()
	}
	cgo.Handle(context.onCall).Delete()
	cgo.Handle(context.pinner).Value().(*runtime.Pinner).Unpin()
	cgo.Handle(context.pinner).Delete()
}

func NewClosure[T any](callback func(T), drop func()) *ClosureContext[T] {
	closure := ClosureContext[T]{}
	closure.onCall = C.uintptr_t(cgo.NewHandle(callback))
	if drop != nil {
		closure.onDrop = C.uintptr_t(cgo.NewHandle(drop))
	}
	context_pinner := &runtime.Pinner{}
	context_pinner.Pin((&closure))
	closure.pinner = C.uintptr_t(cgo.NewHandle(context_pinner))
	return &closure
}
