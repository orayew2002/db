package shared

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

func Marshal(v any) (*structpb.Value, error) {
	return structpb.NewValue(v)
}

func Unmarshal(v *structpb.Value) any {
	return decodeValue(v)
}

// 🔹 MarshalMap: map[string]any -> map[string]*structpb.Value
func MarshalMap(m map[string]any) (map[string]*structpb.Value, error) {
	res := make(map[string]*structpb.Value, len(m))

	for k, v := range m {
		val, err := Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal key %q: %w", k, err)
		}
		res[k] = val
	}

	return res, nil
}

func UnmarshalMap(m map[string]*structpb.Value) map[string]any {
	res := make(map[string]any, len(m))

	for k, v := range m {
		res[k] = Unmarshal(v)
	}

	return res
}

func decodeValue(v *structpb.Value) any {
	if v == nil {
		return nil
	}

	switch kind := v.Kind.(type) {

	case *structpb.Value_NullValue:
		return nil

	case *structpb.Value_NumberValue:
		return kind.NumberValue // always float64

	case *structpb.Value_StringValue:
		return kind.StringValue

	case *structpb.Value_BoolValue:
		return kind.BoolValue

	case *structpb.Value_StructValue:
		return UnmarshalMap(kind.StructValue.Fields)

	case *structpb.Value_ListValue:
		arr := make([]any, len(kind.ListValue.Values))
		for i, item := range kind.ListValue.Values {
			arr[i] = decodeValue(item)
		}
		return arr

	default:
		return nil
	}
}
