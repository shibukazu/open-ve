package validator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

type Validator struct {
	redis  *redis.Client
	logger *slog.Logger
}

func NewValidator(logger *slog.Logger, redis *redis.Client) *Validator {
	return &Validator{redis: redis, logger: logger}
}

func (v *Validator) Validate(id string, variables map[string]interface{}) (bool, string, error) {
	dslVariablesID := dsl.GetVariablesID(id)
	dslVariablesBytes, err := v.redis.Get(dslVariablesID).Bytes()
	if err != nil {
		return false, "", failure.Translate(err, appError.ErrValidateServiceIDNotFound, failure.Messagef("variables not found id: %s", id))
	}
	dslVariables, err := dsl.DeserializeVariables(dslVariablesBytes)
	if err != nil {
		return false, "", err
	}
	celVariables, err := dsl.ToCELVariables(dslVariables)
	if err != nil {
		return false, "", err
	}

	env, err := cel.NewEnv(celVariables...)
	if err != nil {
		return false, "", failure.Translate(err, appError.ErrCELSyantaxError)
	}

	dslAstID := dsl.GetASTID(id)
	encodedAllASTBytes, err := v.redis.Get(dslAstID).Bytes()
	if err != nil {
		return false, "", failure.Translate(err, appError.ErrValidateServiceIDNotFound, failure.Messagef("ast not found id: %s", id))
	}
	allASTBytes, err := decodeAllASTBytes(encodedAllASTBytes)
	if err != nil {
		return false, "", err
	}

	successCh := make(chan bool, len(allASTBytes))
	errorCh := make(chan error, len(allASTBytes))
	failedCELCh := make(chan string, len(allASTBytes))
	// execute all validation asynchronously
	for _, astBytes := range allASTBytes {
		go func(astBytes []byte) {
			var expr exprpb.CheckedExpr
			if err = proto.Unmarshal(astBytes, &expr); err != nil {
				errorCh <- failure.Translate(err, appError.ErrDSLSyntaxError)
				return
			}

			ast := cel.CheckedExprToAst(&expr)
			prg, err := env.Program(ast)
			if err != nil {
				errorCh <- failure.Translate(err, appError.ErrCELSyantaxError)
				return
			}
			res, _, err := prg.Eval(variables)
			if err != nil {
				errorCh <- failure.Translate(err, appError.ErrCELSyantaxError)
				return
			}

			if !res.Value().(bool) {
				failedCEL, err := cel.AstToString(ast)
				if err != nil {
					errorCh <- failure.Translate(err, appError.ErrCELSyantaxError)
					return
				}
				failedCELCh <- failedCEL
				return
			}
			successCh <- true
		}(astBytes)
	}

	result := true
	message := ""
	var failedCELs []string
	for i := 0; i < len(allASTBytes); i++ {
		select {
		case err := <-errorCh:
			return false, "", err
		case failedCEL := <-failedCELCh:
			result = false
			failedCELs = append(failedCELs, failedCEL)
		case <-successCh:
		}
	}
	if len(failedCELs) != 0 {
		message = fmt.Sprintf("failed validations: %s", strings.Join(failedCELs, ", "))
	}
	return result, message, nil
}

func decodeAllASTBytes(bytes []byte) ([][]byte, error) {
	var allASTBytes [][]byte
	if err := json.Unmarshal(bytes, &allASTBytes); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return allASTBytes, nil
}
