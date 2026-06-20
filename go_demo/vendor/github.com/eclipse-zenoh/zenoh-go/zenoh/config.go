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
	"errors"
	"runtime"
	"unsafe"
)

const (
	ConfigModeKey              string = "mode"
	ConfigConnectKey           string = "connect/endpoints"
	ConfigListenKey            string = "listen/endpoints"
	ConfigMulticastScoutingKey string = "scouting/multicast/enabled"
)

// A Zenoh Session config.
type Config struct {
	config *C.z_owned_config_t
}

func configDrop(config *C.z_owned_config_t) {
	C.z_config_drop(C.z_config_move(config))
}

// Create a default configuration.
func NewConfigDefault() Config {
	var c C.z_owned_config_t
	C.z_config_default(&c)
	runtime.SetFinalizer(&c, configDrop)
	return Config{config: &c}
}

// Create a configuration from the JSON5 file.
func NewConfigFromFile(file string) (Config, error) {
	var c C.z_owned_config_t
	fileStr := C.CString(file)
	res := int8(C.zc_config_from_file(&c, fileStr))
	C.free(unsafe.Pointer(fileStr))
	if res == 0 {
		runtime.SetFinalizer(&c, configDrop)
		return Config{config: &c}, nil
	} else {
		return Config{}, newZError(res)
	}
}

// Create a configuration from the JSON5 string.
func NewConfigFromStr(json string) (Config, error) {
	var c C.z_owned_config_t
	jsonStr := C.CString(json)
	res := int8(C.zc_config_from_str(&c, jsonStr))
	C.free(unsafe.Pointer(jsonStr))
	if res == 0 {
		runtime.SetFinalizer(&c, configDrop)
		return Config{config: &c}, nil
	} else {
		return Config{}, newZError(res)
	}
}

// Create a configuration by parsing a file with path stored in ZENOH_CONFIG environment variable.
func NewConfigFromEnv() (Config, error) {
	var c C.z_owned_config_t
	res := int8(C.zc_config_from_env(&c))
	if res == 0 {
		runtime.SetFinalizer(&c, configDrop)
		return Config{config: &c}, nil
	} else {
		return Config{}, newZError(res)
	}
}

// Get config parameter by the string key.
func (config *Config) Get(key string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("config key can not be empty")
	}
	var s C.z_owned_string_t
	loanedConfig := C.z_config_loan(config.config)
	data, size := toDataLen(key)
	res := int8(C.zc_config_get_from_substr(loanedConfig, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(size), &s))
	if res != 0 {
		return "", newZError(res)
	}
	loanedString := C.z_string_loan(&s)
	out := C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))
	C.zc_cgo_string_drop(&s)
	return out, nil
}

// Get the whole config as a JSON string.
func (config *Config) String() string {
	var s C.z_owned_string_t
	loanedConfig := C.z_config_loan(config.config)
	C.zc_config_to_string(loanedConfig, &s)
	loanedString := C.z_string_loan(&s)
	out := C.GoStringN(C.z_string_data(loanedString), C.int(C.z_string_len(loanedString)))
	C.zc_cgo_string_drop(&s)
	return out
}

// Insert a config parameter by the string key.
func (config *Config) InsertJson5(key string, value string) error {
	keyStr := C.CString(key)
	valueStr := C.CString(value)
	loanedConfig := C.z_config_loan_mut(config.config)
	res := int8(C.zc_config_insert_json5(loanedConfig, keyStr, valueStr))
	C.free(unsafe.Pointer(keyStr))
	C.free(unsafe.Pointer(valueStr))
	if res == 0 {
		return nil
	} else {
		return newZError(res)
	}
}

func (config Config) toC() C.z_owned_config_t {
	var out C.z_owned_config_t
	C.z_config_clone(&out, C.z_config_loan(config.config))
	return out
}
