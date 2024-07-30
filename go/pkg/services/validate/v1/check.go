package validatev1

import (
	"context"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	pb "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	variables, err := convertAnyMapToInterfaceMap(req.Variables)
	if err != nil {
		return nil, appError.ToGRPCError(err)
	}
	is_valid, msg, err := s.validator.Validate(req.Id, variables)
	if err != nil {
		return nil, appError.ToGRPCError(err)
	}

	return &pb.CheckResponse{IsValid: is_valid, Message: msg}, nil
}

func convertAnyMapToInterfaceMap(anyMap map[string]*anypb.Any) (map[string]interface{}, error) {
	interfaceMap := make(map[string]interface{})
	for key, anyValue := range anyMap {
		var val interface{}

		// anyValue.TypeUrlで具体的な型を判断してアンマーシャルする
		// TODO: 対応する型を増やす
		switch anyValue.TypeUrl {
		case "type.googleapis.com/google.protobuf.StringValue":
			stringValue := &wrapperspb.StringValue{}
			if err := anyValue.UnmarshalTo(stringValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = stringValue.Value
		case "type.googleapis.com/google.protobuf.DoubleValue":
			doubleValue := &wrapperspb.DoubleValue{}
			if err := anyValue.UnmarshalTo(doubleValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = doubleValue.Value
		case "type.googleapis.com/google.protobuf.FloatValue":
			floatValue := &wrapperspb.FloatValue{}
			if err := anyValue.UnmarshalTo(floatValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = floatValue.Value
		case "type.googleapis.com/google.protobuf.Int32Value":
			intValue := &wrapperspb.Int32Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.Int64Value":
			intValue := &wrapperspb.Int64Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.UInt32Value":
			uintValue := &wrapperspb.UInt32Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.UInt64Value":
			uintValue := &wrapperspb.UInt64Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.SInt32Value":
			intValue := &wrapperspb.Int32Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.SInt64Value":
			intValue := &wrapperspb.Int64Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.Fixed32Value":
			uintValue := &wrapperspb.UInt32Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.Fixed64Value":
			uintValue := &wrapperspb.UInt64Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.BoolValue":
			boolValue := &wrapperspb.BoolValue{}
			if err := anyValue.UnmarshalTo(boolValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = boolValue.Value
		case "type.googleapis.com/google.protobuf.BytesValue":
			bytesValue := &wrapperspb.BytesValue{}
			if err := anyValue.UnmarshalTo(bytesValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = bytesValue
		default:
			return nil, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("unsupported type: %s", anyValue.TypeUrl))
		}

		interfaceMap[key] = val
	}
	return interfaceMap, nil
}
