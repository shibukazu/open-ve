package dsl

import (
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
)

type Result struct {
	ValidationResults []ValidationResult
}

type ValidationResult struct {
	ID              string
	FailedTestCases []string
}

func (d *DSL) Test() (*Result, error) {
	result := &Result{}
	result.ValidationResults = make([]ValidationResult, 0)
	for _, validation := range d.Validations {
		variables := validation.Variables
		celVariables, err := ToCELVariables(variables)
		if err != nil {
			return nil, err
		}
		env, err := cel.NewEnv(celVariables...)
		if err != nil {
			return nil, failure.Translate(err, appError.ErrCELSyantaxError)
		}
		cels := validation.Cels

		failedTestCases := make([]string, 0)
		for _, testCase := range validation.TestCases {
			passAll := true
			for _, cel := range cels {
				ast, issues := env.Compile(cel)
				if issues != nil && issues.Err() != nil {
					return nil, failure.Translate(err, failure.Messagef("Failed to compile CEL: %v", issues.Err()))
				}
				prg, err := env.Program(ast)
				if err != nil {
					return nil, failure.Translate(err, failure.Messagef("Failed to create program: %v", err))
				}
				inputVariables := make(map[string]interface{})
				for _, v := range testCase.Variables {
					inputVariables[v.Name] = v.Value
				}
				res, _, err := prg.Eval(inputVariables)
				if err != nil {
					return nil, failure.Translate(err, failure.Messagef("Failed to evaluate program: %v", err))
				}
				pass, ok := res.Value().(bool)
				if !ok {
					return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("Unsupported result type: %T\nPlease specify one of the following types: bool", res.Value()))
				}
				passAll = passAll && pass
			}
			if passAll != testCase.Expected {
				failedTestCases = append(failedTestCases, testCase.Name)
			}
		}
		validationResult := ValidationResult{
			ID:              validation.ID,
			FailedTestCases: failedTestCases,
		}
		result.ValidationResults = append(result.ValidationResults, validationResult)
	}
	return result, nil
}
