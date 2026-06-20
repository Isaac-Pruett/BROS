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

// The [encoding] of Zenoh data.
//
// [encoding]: https://zenoh.io/docs/manual/abstractions/#encoding
type Encoding struct {
	id     uint16
	schema []byte
}

// Construct default encoding.
func NewEncodingDefault() Encoding {
	return newEncodingFromC(C.zc_internal_encoding_get_data(C.z_encoding_loan_default()))
}

// Construct encoding from string.
func NewEncodinFromString(encoding string) Encoding {
	data, len := toDataLen(encoding)
	var cEncoding C.z_owned_encoding_t
	C.z_encoding_from_substr(&cEncoding, (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len))
	return newEncodingFromOwnedC(&(cEncoding))
}

// Get string representation of encoding.
func (encoding *Encoding) String() string {
	cEncoding := encoding.toCPtr()
	var s C.z_owned_string_t
	C.z_encoding_to_string(C.z_encoding_loan(cEncoding), &s)
	cStringData := C.zc_cgo_string_get_data(C.z_string_loan(&s))
	out := C.GoStringN(cStringData.str_ptr, C.int(cStringData.len))
	C.zc_cgo_string_drop(&s)
	C.zc_cgo_encoding_drop(cEncoding)
	return out
}

// Set schema to this encoding from a string.
//
// Zenoh does not define what a schema is and its semantics is left to the implementer.
// E.g. a common schema for `text/plain` encoding is `utf-8`.
func (encoding *Encoding) SetSchema(schema string) {
	if len(schema) == 0 {
		return
	}
	cEncoding := encoding.toCPtr()
	schemaData, schemaSize := toDataLen(schema)
	loanedEncoding := C.z_encoding_loan(cEncoding)
	C.z_encoding_set_schema_from_substr(loanedEncoding, (*C.char)(unsafe.Pointer(&schemaData[0])), C.size_t(schemaSize))
	*encoding = newEncodingFromOwnedC(cEncoding)
}

func (encoding Encoding) toCData(pinner *runtime.Pinner, out *C.zc_internal_encoding_data_t) {
	out.id = C.uint16_t(encoding.id)
	if len(encoding.schema) > 0 {
		pinner.Pin(&encoding.schema[0])
		out.schema_ptr = (*C.uint8_t)(unsafe.Pointer(&encoding.schema[0]))
		out.schema_len = C.size_t(len(encoding.schema))
	}
}

//go:linkname encodingToUnsafeCData
func encodingToUnsafeCData(encoding Encoding, pinner *runtime.Pinner, out unsafe.Pointer) {
	encoding.toCData(pinner, (*C.zc_internal_encoding_data_t)(out))
}

func (encoding Encoding) toCPtr() *C.z_owned_encoding_t {
	var out C.z_owned_encoding_t
	pinner := runtime.Pinner{}
	var encodingData C.zc_internal_encoding_data_t
	encoding.toCData(&pinner, &encodingData)
	C.zc_internal_encoding_from_data(&out, encodingData)
	pinner.Unpin()
	return &out
}

//go:linkname encodingToUnsafeCPtr
func encodingToUnsafeCPtr(encoding Encoding) unsafe.Pointer {
	return (unsafe.Pointer)(encoding.toCPtr())
}

func newEncodingFromC(cEncoding C.zc_internal_encoding_data_t) Encoding {
	return Encoding{id: uint16(cEncoding.id), schema: C.GoBytes(unsafe.Pointer(cEncoding.schema_ptr), C.int(cEncoding.schema_len))}
}

func newEncodingFromOwnedC(cEncoding *C.z_owned_encoding_t) Encoding {
	e := newEncodingFromC(C.zc_internal_encoding_get_data(C.z_encoding_loan(cEncoding)))
	C.zc_cgo_encoding_drop(cEncoding)
	return e
}

type predefinedEncodings struct{}

// Collection of predefined encodings.
var PredefinedEncodings predefinedEncodings

// Just some bytes.
//
// Constant alias for string: `"zenoh/bytes"`.
//
// This encoding supposes that the payload was created with [NewZBytes]
// similar functions and its data can be accessed via [ZBytes.Bytes].
func (predefinedEncodings) ZenohBytes() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_zenoh_bytes())
	return newEncodingFromC(eData)
}

// A UTF-8 string.
//
// Constant alias for string: `"zenoh/string"`.
//
// This encoding supposes that the payload was created with [NewZBytesFromString]
// similar functions and its data can be accessed via [ZBytes.String].
func (predefinedEncodings) ZenohString() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_zenoh_string())
	return newEncodingFromC(eData)
}

// Zenoh serialized data.
//
// Constant alias for string: `"zenoh/serialized"`.
//
// This encoding supposes that the payload was created with serialization functions.
// The `schema` field may contain the details of serialization format.
func (predefinedEncodings) ZenohSerialized() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_zenoh_serialized())
	return newEncodingFromC(eData)
}

// An application-specific stream of bytes.
//
// Constant alias for string: `"application/octet-stream"`.
func (predefinedEncodings) ApplicationOctetStream() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_octet_stream())
	return newEncodingFromC(eData)
}

// A textual file.
//
// Constant alias for string: `"text/plain"`.
func (predefinedEncodings) TextPlain() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_plain())
	return newEncodingFromC(eData)
}

// JSON data intended to be consumed by an application.
//
// Constant alias for string: `"application/json"`.
func (predefinedEncodings) ApplicationJson() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_json())
	return newEncodingFromC(eData)
}

// JSON data intended to be human readable.
//
// Constant alias for string: `"text/json"`.
func (predefinedEncodings) TextJson() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_json())
	return newEncodingFromC(eData)
}

// A Common Data Representation (CDR)-encoded data.
//
// Constant alias for string: `"application/cdr"`.
func (predefinedEncodings) ApplicationCdr() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_cdr())
	return newEncodingFromC(eData)
}

// A Concise Binary Object Representation (CBOR)-encoded data.
//
// Constant alias for string: `"application/cbor"`.
func (predefinedEncodings) ApplicationCbor() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_cbor())
	return newEncodingFromC(eData)
}

// YAML data intended to be consumed by an application.
//
// Constant alias for string: `"application/yaml"`.
func (predefinedEncodings) ApplicationYaml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_yaml())
	return newEncodingFromC(eData)
}

// YAML data intended to be human readable.
//
// Constant alias for string: `"text/yaml"`.
func (predefinedEncodings) TextYaml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_yaml())
	return newEncodingFromC(eData)
}

// JSON5 encoded data that are human readable.
//
// Constant alias for string: `"text/json5"`.
func (predefinedEncodings) TextJson5() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_json5())
	return newEncodingFromC(eData)
}

// A Python object serialized using [pickle](https://docs.python.org/3/library/pickle.html).
//
// Constant alias for string: `"application/python-serialized-object"`.
func (predefinedEncodings) ApplicationPythonSerializedObject() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_python_serialized_object())
	return newEncodingFromC(eData)
}

// An application-specific protobuf-encoded data.
//
// Constant alias for string: `"application/protobuf"`.
func (predefinedEncodings) ApplicationProtobuf() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_protobuf())
	return newEncodingFromC(eData)
}

// A Java serialized object.
//
// Constant alias for string: `"application/java-serialized-object"`.
func (predefinedEncodings) ApplicationJavaSerializedObject() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_java_serialized_object())
	return newEncodingFromC(eData)
}

// [OpenMetrics](https://github.com/OpenObservability/OpenMetrics) data, commonly used by [Prometheus](https://prometheus.io/).
//
// Constant alias for string: `"application/openmetrics-text"`.
func (predefinedEncodings) ApplicationOpenmetricsText() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_openmetrics_text())
	return newEncodingFromC(eData)
}

// A Portable Network Graphics (PNG) image.
//
// Constant alias for string: `"image/png"`.
func (predefinedEncodings) ImagePng() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_image_png())
	return newEncodingFromC(eData)
}

// A Joint Photographic Experts Group (JPEG) image.
//
// Constant alias for string: `"image/jpeg"`.
func (predefinedEncodings) ImageJpeg() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_image_jpeg())
	return newEncodingFromC(eData)
}

// A Graphics Interchange Format (GIF) image.
//
// Constant alias for string: `"image/gif"`.
func (predefinedEncodings) ImageGif() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_image_gif())
	return newEncodingFromC(eData)
}

// A Bitmap (BMP) image.
//
// Constant alias for string: `"image/bmp"`.
func (predefinedEncodings) ImageBmp() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_image_bmp())
	return newEncodingFromC(eData)
}

// A Web Portable (WebP) image.
//
// Constant alias for string: `"image/webp"`.
func (predefinedEncodings) ImageWebp() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_image_webp())
	return newEncodingFromC(eData)
}

// An XML file intended to be consumed by an application.
//
// Constant alias for string: `"application/xml"`.
func (predefinedEncodings) ApplicationXml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_xml())
	return newEncodingFromC(eData)
}

// An encoded list of tuples, each consisting of a name and a value.
//
// Constant alias for string: `"application/x-www-form-urlencoded"`.
func (predefinedEncodings) ApplicationXWwwFormUrlencoded() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_x_www_form_urlencoded())
	return newEncodingFromC(eData)
}

// An HTML file.
//
// Constant alias for string: `"text/html"`.
func (predefinedEncodings) TextHtml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_html())
	return newEncodingFromC(eData)
}

// An XML file that is human readable.
//
// Constant alias for string: `"text/xml"`.
func (predefinedEncodings) TextXml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_xml())
	return newEncodingFromC(eData)
}

// A CSS file.
//
// Constant alias for string: `"text/css"`.
func (predefinedEncodings) TextCss() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_css())
	return newEncodingFromC(eData)
}

// A JavaScript file.
//
// Constant alias for string: `"text/javascript"`.
func (predefinedEncodings) TextJavascript() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_javascript())
	return newEncodingFromC(eData)
}

// A Markdown file.
//
// Constant alias for string: `"text/markdown"`.
func (predefinedEncodings) TextMarkdown() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_markdown())
	return newEncodingFromC(eData)
}

// A CSV file.
//
// Constant alias for string: `"text/csv"`.
func (predefinedEncodings) TextCsv() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_text_csv())
	return newEncodingFromC(eData)
}

// An application-specific SQL query.
//
// Constant alias for string: `"application/sql"`.
func (predefinedEncodings) ApplicationSql() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_sql())
	return newEncodingFromC(eData)
}

// Constrained Application Protocol (CoAP) data intended for CoAP-to-HTTP and HTTP-to-CoAP proxies.
//
// Constant alias for string: `"application/coap-payload"`.
func (predefinedEncodings) ApplicationCoapPayload() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_coap_payload())
	return newEncodingFromC(eData)
}

// Defines a JSON document structure for expressing a sequence of operations to apply to a JSON document.
//
// Constant alias for string: `"application/json-patch+json"`.
func (predefinedEncodings) ApplicationJsonPatchJson() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_json_patch_json())
	return newEncodingFromC(eData)
}

// A JSON text sequence consists of any number of JSON texts, all encoded in UTF-8.
//
// Constant alias for string: `"application/json-seq"`.
func (predefinedEncodings) ApplicationJsonSeq() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_json_seq())
	return newEncodingFromC(eData)
}

// A JSONPath defines a string syntax for selecting and extracting JSON values from within a given JSON value.
//
// Constant alias for string: `"application/jsonpath"`.
func (predefinedEncodings) ApplicationJsonpath() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_jsonpath())
	return newEncodingFromC(eData)
}

// A JSON Web Token (JWT).
//
// Constant alias for string: `"application/jwt"`.
func (predefinedEncodings) ApplicationJwt() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_jwt())
	return newEncodingFromC(eData)
}

// An application-specific MPEG-4 encoded data, either audio or video.
//
// Constant alias for string: `"application/mp4"`.
func (predefinedEncodings) ApplicationMp4() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_mp4())
	return newEncodingFromC(eData)
}

// A SOAP 1.2 message serialized as XML 1.0.
//
// Constant alias for string: `"application/soap+xml"`.
func (predefinedEncodings) ApplicationSoapXml() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_soap_xml())
	return newEncodingFromC(eData)
}

// A YANG-encoded data commonly used by the Network Configuration Protocol (NETCONF).
//
// Constant alias for string: `"application/yang"`.
func (predefinedEncodings) ApplicationYang() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_application_yang())
	return newEncodingFromC(eData)
}

// A MPEG-4 Advanced Audio Coding (AAC) media.
//
// Constant alias for string: `"audio/aac"`.
func (predefinedEncodings) AudioAac() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_audio_aac())
	return newEncodingFromC(eData)
}

// A Free Lossless Audio Codec (FLAC) media.
//
// Constant alias for string: `"audio/flac"`.
func (predefinedEncodings) AudioFlac() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_audio_flac())
	return newEncodingFromC(eData)
}

// An audio codec defined in MPEG-1, MPEG-2, MPEG-4, or registered at the MP4 registration authority.
//
// Constant alias for string: `"audio/mp4"`.
func (predefinedEncodings) AudioMp4() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_audio_mp4())
	return newEncodingFromC(eData)
}

// An Ogg-encapsulated audio stream.
//
// Constant alias for string: `"audio/ogg"`.
func (predefinedEncodings) AudioOgg() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_audio_ogg())
	return newEncodingFromC(eData)
}

// A Vorbis-encoded audio stream.
//
// Constant alias for string: `"audio/vorbis"`.
func (predefinedEncodings) AudioVorbis() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_audio_vorbis())
	return newEncodingFromC(eData)
}

// A h261-encoded video stream.
//
// Constant alias for string: `"video/h261"`.
func (predefinedEncodings) VideoH261() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_h261())
	return newEncodingFromC(eData)
}

// A h263-encoded video stream.
//
// Constant alias for string: `"video/h263"`.
func (predefinedEncodings) VideoH263() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_h263())
	return newEncodingFromC(eData)
}

// A h264-encoded video stream.
//
// Constant alias for string: `"video/h264"`.
func (predefinedEncodings) VideoH264() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_h264())
	return newEncodingFromC(eData)
}

// A h265-encoded video stream.
//
// Constant alias for string: `"video/h265"`.
func (predefinedEncodings) VideoH265() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_h265())
	return newEncodingFromC(eData)
}

// A h266-encoded video stream.
//
// Constant alias for string: `"video/h266"`.
func (predefinedEncodings) VideoH266() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_h266())
	return newEncodingFromC(eData)
}

// A video codec defined in MPEG-1, MPEG-2, MPEG-4, or registered at the MP4 registration authority.
//
// Constant alias for string: `"video/mp4"`.
func (predefinedEncodings) VideoMp4() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_mp4())
	return newEncodingFromC(eData)
}

// An Ogg-encapsulated video stream.
//
// Constant alias for string: `"video/ogg"`.
func (predefinedEncodings) VideoOgg() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_ogg())
	return newEncodingFromC(eData)
}

// An uncompressed, studio-quality video stream.
//
// Constant alias for string: `"video/raw"`.
func (predefinedEncodings) VideoRaw() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_raw())
	return newEncodingFromC(eData)
}

// A VP8-encoded video stream.
//
// Constant alias for string: `"video/vp8"`.
func (predefinedEncodings) VideoVp8() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_vp8())
	return newEncodingFromC(eData)
}

// A VP9-encoded video stream.
//
// Constant alias for string: `"video/vp9"`.
func (predefinedEncodings) VideoVp9() Encoding {
	eData := C.zc_internal_encoding_get_data(C.z_encoding_video_vp9())
	return newEncodingFromC(eData)
}
