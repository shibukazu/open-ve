package validator

import (
	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
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

func (v *Validator) Validate(id string, variables map[string]interface{}) (bool, error) {
	dslVariablesID := dsl.GetVariablesID(id)
	dslVariablesBytes, err := v.redis.Get(dslVariablesID).Bytes()
	if err != nil {
		return false, failure.Translate(err, appError.ErrValidateServiceIDNotFound, failure.Messagef("variables not found id: %s", id))
	}
	dslVariables, err := dsl.DeserializeVariables(dslVariablesBytes)
	if err != nil {
		return false, err
	}
	celVariables, err := dsl.ToCELVariables(dslVariables)
	if err != nil {
		return false, err
	}

	env, err := cel.NewEnv(celVariables...)
	if err != nil {
		return false, failure.Translate(err, appError.ErrCELSyantaxError)
	}

	dslAstID := dsl.GetASTID(id)
	dslAstBytes, err := v.redis.Get(dslAstID).Bytes()
	if err != nil {
		return false, failure.Translate(err, appError.ErrValidateServiceIDNotFound, failure.Messagef("ast not found id: %s", id))
	}

	var expr exprpb.CheckedExpr
	if err = proto.Unmarshal(dslAstBytes, &expr); err != nil {
		return false, failure.Translate(err, appError.ErrCELSyantaxError)
	}

	ast := cel.CheckedExprToAst(&expr)
	prg, err := env.Program(ast)
	if err != nil {
		return false, failure.Translate(err, appError.ErrCELSyantaxError)
	}
	res, _, err := prg.Eval(variables)
	if err != nil {
		return false, failure.Translate(err, appError.ErrCELSyantaxError)
	}

	return res.Value().(bool), nil
}
