package shared

import "google.golang.org/protobuf/types/known/structpb"

func Marshal(m map[string]any) (map[string]*structpb.Value, error) {
	res := make(map[string]*structpb.Value, len(m))

	for k, v := range m {
		val, err := structpb.NewValue(v)
		if err != nil {
			return nil, err
		}
		res[k] = val
	}

	return res, nil
}

func Unmarshal(m map[string]*structpb.Value) map[string]any {
	res := make(map[string]any, len(m))

	for k, v := range m {
		res[k] = decodeValue(v)
	}

	return res
}

func decodeValue(v *structpb.Value) any {
	switch kind := v.Kind.(type) {

	case *structpb.Value_NullValue:
		return nil

	case *structpb.Value_NumberValue:
		return kind.NumberValue

	case *structpb.Value_StringValue:
		return kind.StringValue

	case *structpb.Value_BoolValue:
		return kind.BoolValue

	case *structpb.Value_StructValue:
		return Unmarshal(kind.StructValue.Fields)

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
