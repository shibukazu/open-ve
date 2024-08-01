package validator

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/store"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
)

type Validator struct {
	store  store.Store
	logger *slog.Logger
}

func NewValidator(logger *slog.Logger, store store.Store) *Validator {
	return &Validator{logger: logger, store: store}
}

func (v *Validator) Validate(id string, variables map[string]interface{}) (bool, string, error) {
	dslVariables, err := v.store.ReadVariables(id)
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

	allEncodedAST, err := v.store.ReadAllEncodedAST(id)
	if err != nil {
		return false, "", err
	}

	successCh := make(chan bool, len(allEncodedAST))
	errorCh := make(chan error, len(allEncodedAST))
	failedCELCh := make(chan string, len(allEncodedAST))
	// execute all validation asynchronously
	for _, encodedAST := range allEncodedAST {
		go func(encodedAST []byte) {
			var expr exprpb.CheckedExpr
			if err = proto.Unmarshal(encodedAST, &expr); err != nil {
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
		}(encodedAST)
	}

	result := true
	message := ""
	var failedCELs []string
	for i := 0; i < len(allEncodedAST); i++ {
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
