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
	is_valid, err := s.validator.Validate(req.Id, variables)
	if err != nil {
		return nil, appError.ToGRPCError(err)
	}

	return &pb.CheckResponse{IsValid: is_valid}, nil
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
		case "type.googleapis.com/google.protobuf.Int32Value":
			intValue := &wrapperspb.Int32Value{}
			if err := anyValue.UnmarshalTo(intValue); err != nil {
				return nil, failure.Translate(err, appError.ErrRequestParameterInvalid)
			}
			val = intValue.Value
		default:
			return nil, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("unsupported type: %s", anyValue.TypeUrl))
		}

		interfaceMap[key] = val
	}
	return interfaceMap, nil
}