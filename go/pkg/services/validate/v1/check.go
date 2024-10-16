package validatev1

import (
	"context"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	pb "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Service) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	results := make([]*pb.ValidationResult, 0, len(req.Validations))

	for _, validation := range req.Validations {
		variables, err := convertAnyMapToInterfaceMap(validation.Variables)
		if err != nil {
			logger.LogError(s.logger, err)
			return nil, appError.ToGRPCError(err)
		}
		is_valid, msg, err := s.validator.Validate(validation.Id, variables)
		if err != nil {
			logger.LogError(s.logger, err)
			return nil, appError.ToGRPCError(err)
		}
		results = append(results, &pb.ValidationResult{Id: validation.Id, IsValid: is_valid, Message: msg})
	}

	return &pb.CheckResponse{Results: results}, nil
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
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal string value"))
			}
			val = stringValue.Value
		case "type.googleapis.com/google.protobuf.DoubleValue":
			doubleValue := &wrapperspb.DoubleValue{}
			if err := anyValue.UnmarshalTo(doubleValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal double value"))
			}
			val = doubleValue.Value
		case "type.googleapis.com/google.protobuf.FloatValue":
			floatValue := &wrapperspb.FloatValue{}
			if err := anyValue.UnmarshalTo(floatValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal float value"))
			}
			val = floatValue.Value
		case "type.googleapis.com/google.protobuf.Int32Value":
			intValue := &wrapperspb.Int32Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal int32 value"))
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.Int64Value":
			intValue := &wrapperspb.Int64Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal int64 value"))
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.UInt32Value":
			uintValue := &wrapperspb.UInt32Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal uint32 value"))
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.UInt64Value":
			uintValue := &wrapperspb.UInt64Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal uint64 value"))
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.SInt32Value":
			intValue := &wrapperspb.Int32Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal sint32 value"))
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.SInt64Value":
			intValue := &wrapperspb.Int64Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal sint64 value"))
			}
			val = intValue.Value
		case "type.googleapis.com/google.protobuf.Fixed32Value":
			uintValue := &wrapperspb.UInt32Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal fixed32 value"))
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.Fixed64Value":
			uintValue := &wrapperspb.UInt64Value{}
			if err := anyValue.UnmarshalTo(uintValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal fixed64 value"))
			}
			val = uintValue.Value
		case "type.googleapis.com/google.protobuf.BoolValue":
			boolValue := &wrapperspb.BoolValue{}
			if err := anyValue.UnmarshalTo(boolValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal bool value"))
			}
			val = boolValue.Value
		case "type.googleapis.com/google.protobuf.BytesValue":
			bytesValue := &wrapperspb.BytesValue{}
			if err := anyValue.UnmarshalTo(bytesValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to unmarshal bytes value"))
			}
			val = bytesValue
		default:
			return nil, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("unsupported type: %s", anyValue.TypeUrl))
		}

		interfaceMap[key] = val
	}
	return interfaceMap, nil
}
