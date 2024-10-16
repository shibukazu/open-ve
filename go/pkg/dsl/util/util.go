package util

import (
	"io"
	"os"

	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"gopkg.in/yaml.v3"
)

func DSLVariableToCELVariable(v *dsl.Variable) (cel.EnvOption, error) {
	switch v.Type {
	case "int":
		return cel.Variable(v.Name, cel.IntType), nil
	case "uint":
		return cel.Variable(v.Name, cel.UintType), nil
	case "double":
		return cel.Variable(v.Name, cel.DoubleType), nil
	case "bool":
		return cel.Variable(v.Name, cel.BoolType), nil
	case "bytes":
		return cel.Variable(v.Name, cel.BytesType), nil
	case "string":
		return cel.Variable(v.Name, cel.StringType), nil
	// TODO: listとmap向けの再帰パースの実装
	default:
		return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("Unsupported variable type: %s\nPlease specify one of the following types: int, uint, double, bool, string, bytes", v.Type))
	}
}

func DSLVariablesToCELVariables(vars []dsl.Variable) ([]cel.EnvOption, error) {
	celVars := make([]cel.EnvOption, 0, len(vars))
	for _, v := range vars {
		v, err := DSLVariableToCELVariable(&v)
		if err != nil {
			return nil, err
		}
		celVars = append(celVars, v)
	}
	return celVars, nil
}

func ParseDSLYAML(yamlFilePath string) (*dsl.DSL, error) {
	yamlFile, err := os.Open(yamlFilePath)
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()

	yamlBytes, err := io.ReadAll(yamlFile)
	if err != nil {
		return nil, err
	}

	dsl := &dsl.DSL{}
	err = yaml.Unmarshal(yamlBytes, &dsl)
	if err != nil {
		return nil, err
	}

	return dsl, nil
}
