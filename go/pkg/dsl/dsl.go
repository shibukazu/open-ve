package dsl

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

type DSLReader struct {
	redis *redis.Client
}

type DSL struct {
	Validations []Validation `yaml:"validations"`
}

type Validation struct {
	ID        string     `yaml:"id"`
	Cel       string     `yaml:"cel"`
	Variables []Variable `yaml:"variables"`
}

type Variable struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
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
		vars := make([]cel.EnvOption, 0, len(v.Variables))
		for _, vv := range v.Variables {
			v, err := r.parseVariable(vv)
			if err != nil {
				return err
			}
			vars = append(vars, v)
		}

		env, err := cel.NewEnv(vars...)
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
		if err := r.redis.Set(v.ID, astBytes, 0).Err(); err != nil {
			log.Fatalf("Failed to save AST to Redis: %v", err)
		}
	}

	return nil
}

func (r *DSLReader) parseVariable(v Variable) (cel.EnvOption, error) {
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
