package dsl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

type Variable struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"`
}

func (v *Variable) parseVariable() (cel.EnvOption, error) {
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
	// TODO: listとmap向けの再帰パースの実装

	default:
		return nil, fmt.Errorf("unknown type: %s", v.Type)
	}
}

type Validation struct {
	ID        string     `yaml:"id"`
	Cel       string     `yaml:"cel"`
	Variables []Variable `yaml:"variables" json:"variables"`
}

func GetASTID(id string) string {
	return fmt.Sprintf("ast:%s", id)
}
func GetVariablesID(id string) string {
	return fmt.Sprintf("vars:%s", id)
}

func (v *Validation) serializeVariables() ([]byte, error) {
	bytes, err := json.Marshal(v.Variables)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func DeserializeVariables(bytes []byte) ([]Variable, error) {
	var vars []Variable
	err := json.Unmarshal(bytes, &vars)
	if err != nil {
		return nil, err
	}
	return vars, nil
}

func ToCELVariables(vars []Variable) ([]cel.EnvOption, error) {
	celVars := make([]cel.EnvOption, 0, len(vars))
	for _, v := range vars {
		v, err := v.parseVariable()
		if err != nil {
			return nil, err
		}
		celVars = append(celVars, v)
	}
	return celVars, nil
}

type DSL struct {
	Validations []Validation `yaml:"validations"`
}

type DSLReader struct {
	redis *redis.Client
}

func NewDSLReader(redis *redis.Client) *DSLReader {
	return &DSLReader{redis: redis}
}

func (r *DSLReader) Read(ctx context.Context, bytes []byte) error {
	var dsl DSL
	err := yaml.Unmarshal(bytes, &dsl)
	if err != nil {
		log.Fatal(err)
	}
	err = r.parseAndSaveDSL(dsl)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (r *DSLReader) parseAndSaveDSL(dsl DSL) error {
	for _, v := range dsl.Validations {
		// Save Variables to Redis
		variablesBytes, err := v.serializeVariables()
		if err != nil {
			log.Fatalf("Failed to serialize Variables: %v", err)
		}
		if err := r.redis.Set(GetVariablesID(v.ID), variablesBytes, 0).Err(); err != nil {
			log.Fatalf("Failed to save Variables to Redis: %v", err)
		}

		celVariables, err := ToCELVariables(v.Variables)
		if err != nil {
			log.Fatalf("Failed to convert Variables to CEL Variables: %v", err)
		}
		env, err := cel.NewEnv(celVariables...)
		if err != nil {
			return err
		}
		ast, issues := env.Compile(v.Cel)
		if issues != nil && issues.Err() != nil {
			return issues.Err()
		}
		// Convert AST to Proto
		expr, err := cel.AstToCheckedExpr(ast)
		if err != nil {
			return err
		}
		astBytes, err := proto.Marshal(expr)
		if err != nil {
			log.Fatalf("Failed to serialize AST: %v", err)
		}
		// Save AST to Redis
		if err := r.redis.Set(GetASTID(v.ID), astBytes, 0).Err(); err != nil {
			log.Fatalf("Failed to save AST to Redis: %v", err)
		}
	}

	return nil
}
