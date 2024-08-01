package dsl

import (
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
)

type Variable struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
}

func (v *Variable) ParseVariable() (cel.EnvOption, error) {
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

func ToCELVariables(vars []Variable) ([]cel.EnvOption, error) {
	celVars := make([]cel.EnvOption, 0, len(vars))
	for _, v := range vars {
		v, err := v.ParseVariable()
		if err != nil {
			return nil, err
		}
		celVars = append(celVars, v)
	}
	return celVars, nil
}

type Validation struct {
	ID        string     `yaml:"id" json:"id"`
	Cels      []string   `yaml:"cels" json:"cels"`
	Variables []Variable `yaml:"variables" json:"variables"`
}

type DSL struct {
	Validations []Validation `yaml:"validations" json:"validations"`
}
