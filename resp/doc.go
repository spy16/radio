// Package resp provides functions and types for RESP protocol parsing.
// See https://redis.io/topics/protocol for protocol specification.
//
// In addition to the standard protocol, this package also supports
// optional extensions such as support for 64-bit signed float typed
// values in a non-ambiguous way.
package resp
