// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import "errors"

var (
	// ErrS1 is a static error if templates encoded in the concrete XML syntax are in
	// fact not well-formed, do not follow the rules of XML namespaces or are invalid
	// with respect to the schema in Appendix 1 in FAST 1.1 specification.
	ErrS1 = errors.New("static error: S1")

	// ErrS2 is a static error if an operator is specified for a field of a type to
	// which the operator is not applicable.
	ErrS2 = errors.New("static error: S2")

	// ErrS3 is a static error if an initial value specified by the value attribute
	// in the concrete syntax can not be converted to a value of the type of the field.
	ErrS3 = errors.New("static error: S3")

	// ErrS4 is a static error if no initial value is specified for a constant operator.
	ErrS4 = errors.New("static error: S4")

	// ErrS5 is a static error if no initial value is specified for a default operator
	// on a mandatory field.
	ErrS5 = errors.New("static error: S5")

	// ErrD1 is a dynamic error if type of a field in a template can not be converted
	// to or from the type of the corresponding application field.
	ErrD1 = errors.New("dynamic error: D1")

	// ErrD2 is a dynamic error if an integer in the stream does not fall within the
	// bounds of the specifies integer type specified on the corresponding field.
	ErrD2 = errors.New("dynamic error: D2")

	// ErrD3 is a dynamic error if a decimal value can not be encoded due to limitation
	// introduced by using individual operators on exponent and mantissa.
	ErrD3 = errors.New("dynamic error: D3")

	// ErrD4 is a dynamic error if the type of the previous value is not the same as
	// the type of the field of the current operator.
	ErrD4 = errors.New("dynamic error: D4")

	// ErrD5 is a dynamic error if a mandatory field is not present in the stream, has
	// an undefined previous value and there is no initial value in the instruction
	// context.
	ErrD5 = errors.New("dynamic error: D5")

	// ErrD6 is a dynamic error if a mandatory field is not present in the stream and
	// has an empty previous value.
	ErrD6 = errors.New("dynamic error: D6")

	// ErrD7 is a dynamic error if the subtraction length exceeds the length of the base
	// value or if it does not fall in the value rang of an int32.
	ErrD7 = errors.New("dynamic error: D7")

	// ErrD8 is a dynamic error if the name specified on a static template reference
	// does not point to a template known be the encoder or decoder.
	ErrD8 = errors.New("dynamic error: D8")

	// ErrD9 is a dynamic error if a decoder can not find a template associated with a
	// template identifier appearing in the stream.
	ErrD9 = errors.New("dynamic error: D9")

	// ErrD10 is a dynamic error to convert byte vectors to and from other types
	// than string.
	ErrD10 = errors.New("dynamic error: D10")

	// ErrD11 is a dynamic error if the syntax of a string does not follow the rules for
	// the type converted to.
	ErrD11 = errors.New("dynamic error: D11")

	// ErrD12 is a dynamic error if a block length preamble is zero.
	ErrD12 = errors.New("dynamic error: D12")

	// ErrR1 is a reportable error if a decimal can not be represented by an exponent in
	// the range [-63 ... 63] of if the mantissa does not fit in an int64.
	ErrR1 = errors.New("reportable error: R1")

	// ErrR2 is a reportable error if the combined value after applying a tail or delta
	// operator to a Unicode string is not a valid UTF-8 sequence.
	ErrR2 = errors.New("reportable error: R2")

	// ErrR3 is a reportable error if a Unicode string that is being converted to an
	// ASCII string contains characters that are outside the ASCII character set.
	ErrR3 = errors.New("reportable error: R3")

	// ErrR4 is a reportable error if a value of an integer cannot be represented in the
	// target integer type in a conversion.
	ErrR4 = errors.New("reportable error: R4")

	// ErrR5 is a reportable error if a decimal being converted to an integer has a
	// negative exponent or if the resulting integer does not fit the target integer
	// type.
	ErrR5 = errors.New("reportable error: R5")

	// ErrR6 is a reportable error if an integer appears in an overlong encoding.
	ErrR6 = errors.New("reportable error: R6")

	// ErrR7 is a reportable error if a presence map is overlong.
	ErrR7 = errors.New("reportable error: R7")

	// ErrR8 is a reportable error if a presence map contains more bits than required.
	ErrR8 = errors.New("reportable error: R8")

	// ErrR9 is a reportable error if a string appears in an overlong encoding.
	ErrR9 = errors.New("reportable error: R9")
)
