package tester

import (
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/dsl/util"
)

type Result struct {
	ValidationResults []ValidationResult
}

type ValidationResult struct {
	ID               string
	FailedTestCases  []string
	TestCaseNotFound bool
}

func TestDSL(d *dsl.DSL) (*Result, error) {
	result := &Result{}
	result.ValidationResults = make([]ValidationResult, 0)
	for _, validation := range d.Validations {
		if len(validation.TestCases) == 0 {
			result.ValidationResults = append(result.ValidationResults, ValidationResult{
				ID:               validation.ID,
				FailedTestCases:  []string{},
				TestCaseNotFound: true,
			})
			continue
		}
		variables := validation.Variables
		celVariables, err := util.DSLVariablesToCELVariables(variables)
		if err != nil {
			return nil, err
		}
		env, err := cel.NewEnv(celVariables...)
		if err != nil {
			return nil, failure.Translate(err, appError.ErrDSLSyntaxError, failure.Messagef("failed to create CEL environment: %v", err))
		}
		cels := validation.Cels

		failedTestCases := make([]string, 0)
		for _, testCase := range validation.TestCases {
			passAll := true
			for _, cel := range cels {
				ast, issues := env.Compile(cel)
				if issues != nil && issues.Err() != nil {
					return nil, failure.Translate(err, failure.Messagef("failed to compile CEL: %v", issues.Err()))
				}
				prg, err := env.Program(ast)
				if err != nil {
					return nil, failure.Translate(err, failure.Messagef("failed to create program: %v", err))
				}
				inputVariables := make(map[string]interface{})
				for _, v := range testCase.Variables {
					inputVariables[v.Name] = v.Value
				}
				res, _, err := prg.Eval(inputVariables)
				if err != nil {
					return nil, failure.Translate(err, failure.Messagef("failed to evaluate program: %v", err))
				}
				pass, ok := res.Value().(bool)
				if !ok {
					return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("unsupported result type: %T\nplease specify one of the following types: bool", res.Value()))
				}
				passAll = passAll && pass
			}
			if passAll != testCase.Expected {
				failedTestCases = append(failedTestCases, testCase.Name)
			}
		}
		validationResult := ValidationResult{
			ID:               validation.ID,
			FailedTestCases:  failedTestCases,
			TestCaseNotFound: false,
		}
		result.ValidationResults = append(result.ValidationResults, validationResult)
	}
	return result, nil
}
