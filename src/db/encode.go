package db

import (
	"encoding/json"
	"strconv"
)

// appendMapJSON encodes m as JSON and appends to dst with zero heap allocations
// for the common value types (string, int, float64, bool, nil).
func appendMapJSON(dst []byte, m map[string]any) []byte {
	dst = append(dst, '{')
	first := true
	for k, v := range m {
		if !first {
			dst = append(dst, ',')
		}
		first = false
		dst = append(dst, '"')
		dst = appendEscapedStr(dst, k)
		dst = append(dst, '"', ':')
		dst = appendAnyJSON(dst, v)
	}
	return append(dst, '}')
}

func appendAnyJSON(dst []byte, v any) []byte {
	switch x := v.(type) {
	case string:
		dst = append(dst, '"')
		dst = appendEscapedStr(dst, x)
		return append(dst, '"')
	case int:
		return strconv.AppendInt(dst, int64(x), 10)
	case int32:
		return strconv.AppendInt(dst, int64(x), 10)
	case int64:
		return strconv.AppendInt(dst, x, 10)
	case uint:
		return strconv.AppendUint(dst, uint64(x), 10)
	case uint64:
		return strconv.AppendUint(dst, x, 10)
	case float32:
		return strconv.AppendFloat(dst, float64(x), 'f', -1, 32)
	case float64:
		return strconv.AppendFloat(dst, x, 'f', -1, 64)
	case bool:
		if x {
			return append(dst, "true"...)
		}
		return append(dst, "false"...)
	case nil:
		return append(dst, "null"...)
	default:
		b, _ := json.Marshal(x)
		return append(dst, b...)
	}
}

// appendEscapedStr appends s to dst with JSON string escaping, no allocations.
func appendEscapedStr(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			dst = append(dst, '\\', '"')
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if c < 0x20 {
				dst = append(dst, '\\', 'u', '0', '0', hexNibble(c>>4), hexNibble(c&0xf))
			} else {
				dst = append(dst, c)
			}
		}
	}
	return dst
}

func hexNibble(c byte) byte {
	if c < 10 {
		return '0' + c
	}
	return 'a' + c - 10
}
