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

// A Zenoh message handler interface.
//
// ToCbDropHandler must turn a handler into a triplet, consisting of
//   - a non-nil message processing function which is called every time a Zenoh entity receives a message,
//   - a drop function that is called when Zenoh entity is destroyed or GET request is finalized (can be nil),
//   - a receive channel which exposes data, fed by message processing function (can be nil).
//
// Zenoh provides [Closure], [FifoChannel] and [RingChannel] handler implementations.
type Handler[T any] interface {
	ToCbDropHandler() (func(T), func(), <-chan T)
}

// A closure is a [Handler] that contains all the elements for stateful, memory-leak-free Zenoh entities callbacks.
//
// Closures are not guaranteed not to be called concurrently.
// It is guaranteed that:
//   - `Call` will never be called once `drop` has started.
//   - `Drop` will only be called **once**, and **after every** `Call` has ended.
//   - The two previous guarantees imply that `Call` and `Drop` are never called concurrently.
type Closure[T any] struct {
	Call func(T) // A non-nil function that is called every time a Zenoh entity receives a new message.
	Drop func()  // A function that is called when Zenoh entity is destroyed or GET request is finalized. Can be nil.
}

// A FIFO channel. A [Handler] that exposes Zenoh messages in FIFO order, will block Zenoh callback execution when full.
type FifoChannel[T any] struct {
	channel chan T
}

// Create a [FifoChannel] with specified capacity.
func NewFifoChannel[T any](capacity int) FifoChannel[T] {
	return FifoChannel[T]{channel: make(chan T, capacity)}
}

// A Ring channel. A [Handler] that exposes Zenoh message in a FIFO order. When full, the older unprocessed messages will be removed to leave room for the new ones.
type RingChannel[T any] struct {
	channel chan T
}

// Create a [RingChannel]  with specified capacity.
func NewRingChannel[T any](capacity int) RingChannel[T] {
	return RingChannel[T]{channel: make(chan T, capacity)}
}

func (closure Closure[T]) ToCbDropHandler() (func(T), func(), <-chan T) {
	return closure.Call, closure.Drop, nil
}

func (channel FifoChannel[T]) ToCbDropHandler() (func(T), func(), <-chan T) {
	call := func(t T) {
		channel.channel <- t
	}
	drop := func() {
		close(channel.channel)
	}
	return call, drop, channel.channel
}

func (channel RingChannel[T]) ToCbDropHandler() (func(T), func(), <-chan T) {
	call := func(t T) {
		select {
		case channel.channel <- t:
		default:
			<-channel.channel
			channel.channel <- t
		}
	}
	drop := func() {
		close(channel.channel)
	}
	return call, drop, channel.channel
}
