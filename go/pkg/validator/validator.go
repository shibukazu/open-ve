package validator

import (
	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

type Validator struct {
	redis *redis.Client
}

func NewValidator(redis *redis.Client) *Validator {
	return &Validator{redis: redis}
}

func (v *Validator) Validate(id string, values map[string]interface{}) (bool, error) {
	variablesID := dsl.GetVariablesID(id)
	variablesBytes, err := v.redis.Get(variablesID).Bytes()
	if err != nil {
		return false, err
	}
	variables, err := dsl.DeserializeVariables(variablesBytes)
	if err != nil {
		return false, err
	}
	celVariables, err := dsl.ToCELVariables(variables)
	if err != nil {
		return false, err
	}

	env, err := cel.NewEnv(celVariables...)
	if err != nil {
		return false, err
	}

	astID := dsl.GetASTID(id)
	astBytes, err := v.redis.Get(astID).Bytes()
	if err != nil {
		return false, err
	}

	var expr exprpb.CheckedExpr
	if err = proto.Unmarshal(astBytes, &expr); err != nil {
		return false, err
	}

	ast := cel.CheckedExprToAst(&expr)
	prg, err := env.Program(ast)
	if err != nil {
		return false, err
	}
	res, _, err := prg.Eval(values)
	if err != nil {
		return false, err
	}

	return res.Value().(bool), nil
}
