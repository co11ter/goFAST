// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

const (
	uintSize = 32 << (^uint(0) >> 32 & 1)
	maxInt = 1<<(uintSize-1) - 1
	maxUint = 1<<uintSize - 1
)

func castTo(src, dst interface{}) error {
	switch src.(type) {
	case []byte:
		return castByteVectorTo(src.([]byte), dst)
	case string:
		return castStringTo(src.(string), dst)
	case uint32:
		return castUintTo(uint64(src.(uint32)), dst)
	case uint64:
		return castUintTo(src.(uint64), dst)
	case int32:
		return castIntTo(int64(src.(int32)), dst)
	case int64:
		return castIntTo(src.(int64), dst)
	case float64:
		return castFloatTo(src.(float64), dst)
	}

	return nil
}

func castByteVectorTo(src []byte, dst interface{}) (err error) {
	switch dst.(type) {
	case *string:
		*dst.(*string) = string(src)
	default:
		err = ErrD10
	}

	return
}

func castStringTo(src string, dst interface{}) (err error) {
	src = strings.TrimSpace(src)
	switch dst.(type) {
	case *int:
		var tmp int64
		tmp, err = strconv.ParseInt(src, 10, 0)
		*dst.(*int) = int(tmp)
	case *int64:
		*dst.(*int64), err = strconv.ParseInt(src, 10, 0)
	case *int32:
		var tmp int64
		tmp, err = strconv.ParseInt(src, 10, 32)
		*dst.(*int32) = int32(tmp)
	case *int16:
		var tmp int64
		tmp, err = strconv.ParseInt(src, 10, 16)
		*dst.(*int16) = int16(tmp)
	case *int8:
		var tmp int64
		tmp, err = strconv.ParseInt(src, 10, 8)
		*dst.(*int8) = int8(tmp)
	case *uint:
		var tmp uint64
		tmp, err = strconv.ParseUint(src, 10, 0)
		*dst.(*uint) = uint(tmp)
	case *uint64:
		*dst.(*uint64), err = strconv.ParseUint(src, 10, 0)
	case *uint32:
		var tmp uint64
		tmp, err = strconv.ParseUint(src, 10, 32)
		*dst.(*uint32) = uint32(tmp)
	case *uint16:
		var tmp uint64
		tmp, err = strconv.ParseUint(src, 10, 16)
		*dst.(*uint16) = uint16(tmp)
	case *uint8:
		var tmp uint64
		tmp, err = strconv.ParseUint(src, 10, 8)
		*dst.(*uint8) = uint8(tmp)
	case *float64:
		*dst.(*float64), err = strconv.ParseFloat(src, 0)
	case *float32:
		var tmp float64
		tmp, err = strconv.ParseFloat(src, 32)
		*dst.(*float32) = float32(tmp)
	case *[]byte:
		*dst.(*[]byte) = []byte(src)
	case *string:
		*dst.(*string) = src
	}

	checkStringErr(err, dst)
	return
}

func castUintTo(src uint64, dst interface{}) (err error) {
	switch dst.(type) {
	case *int:
		*dst.(*int) = int(src)
		if src > maxInt {
			err = ErrR4
		}
	case *int64:
		*dst.(*int64) = int64(src)
		if src > math.MaxInt64 {
			err = ErrR4
		}
	case *int32:
		*dst.(*int32) = int32(src)
		if src > math.MaxInt32 {
			err = ErrR4
		}
	case *int16:
		*dst.(*int16) = int16(src)
		if src > math.MaxInt16 {
			err = ErrR4
		}
	case *int8:
		*dst.(*int8) = int8(src)
		if src > math.MaxInt8 {
			err = ErrR4
		}
	case *uint:
		*dst.(*uint) = uint(src)
		if src > maxUint {
			err = ErrR4
		}
	case *uint64:
		*dst.(*uint64) = src
	case *uint32:
		*dst.(*uint32) = uint32(src)
		if src > math.MaxUint32 {
			err = ErrR4
		}
	case *uint16:
		*dst.(*uint16) = uint16(src)
		if src > math.MaxUint16 {
			err = ErrR4
		}
	case *uint8:
		*dst.(*uint8) = uint8(src)
		if src > math.MaxUint8 {
			err = ErrR4
		}
	case *float64:
		*dst.(*float64) = float64(src)
	case *float32:
		*dst.(*float32) = float32(src)
	case *[]byte:
		err = ErrD10
	case *string:
		*dst.(*string) = strconv.FormatUint(src, 10)
	}
	return
}

func castIntTo(src int64, dst interface{}) (err error) {
	switch dst.(type) {
	case *int:
		*dst.(*int) = int(src)
		if src > maxInt {
			err = ErrR4
		}
	case *int64:
		*dst.(*int64) = src
	case *int32:
		*dst.(*int32) = int32(src)
		if src > math.MaxInt32 {
			err = ErrR4
		}
	case *int16:
		*dst.(*int16) = int16(src)
		if src > math.MaxInt16 {
			err = ErrR4
		}
	case *int8:
		*dst.(*int8) = int8(src)
		if src > math.MaxInt8 {
			err = ErrR4
		}
	case *uint:
		*dst.(*uint) = uint(src)
		if uint(src) > maxUint {
			err = ErrR4
		}
	case *uint64:
		*dst.(*uint64) = uint64(src)
	case *uint32:
		*dst.(*uint32) = uint32(src)
		if src > math.MaxUint32 {
			err = ErrR4
		}
	case *uint16:
		*dst.(*uint16) = uint16(src)
		if src > math.MaxUint16 {
			err = ErrR4
		}
	case *uint8:
		*dst.(*uint8) = uint8(src)
		if src > math.MaxUint8 {
			err = ErrR4
		}
	case *float64:
		*dst.(*float64) = float64(src)
	case *float32:
		*dst.(*float32) = float32(src)
	case *[]byte:
		err = ErrD10
	case *string:
		*dst.(*string) = strconv.FormatInt(src, 10)
	}
	return
}

func castFloatTo(src float64, dst interface{}) (err error) {
	switch dst.(type) {
	case *int, *int64, *int32, *int16, *int8, *uint, *uint64, *uint32, *uint16, *uint8:
		err = checkFloatErr(src)
		if err != nil {
			return
		}
		err = castUintTo(uint64(src), dst)
	case *float64:
		*dst.(*float64) = src
	case *float32:
		*dst.(*float32) = float32(src)
	case *[]byte:
		err = ErrD10
	case *string:
		*dst.(*string) = strconv.FormatFloat(src, 'f', -1, 64)
	}
	return
}

func castStringToASCII(src interface{}, dst *string) (err error) {
	err = castTo(src, dst)
	if err != nil {
		return
	}
	for i := 0; i < len(*dst); i++ {
		if (*dst)[i] > unicode.MaxASCII {
			return ErrR3
		}
	}
	return
}

func checkFloatErr(src float64) error {
	if expDecimal(src) > 0 {
		return ErrR5
	}
	return nil
}

func checkStringErr(err error, dst interface{}) {
	switch dst.(type) {
	case *int, *int64, *int32, *int16, *int8, *uint, *uint64, *uint32, *uint16, *uint8:
		if e, ok := err.(*strconv.NumError); ok {
			if e.Err == strconv.ErrRange {
				err = ErrR4
			} else {
				err = ErrD11
			}
		}
	case *float32, *float64:
		if e, ok := err.(*strconv.NumError); ok {
			if e.Err == strconv.ErrRange {
				err = ErrR1
			} else {
				err = ErrD11
			}
		}
	}
}
