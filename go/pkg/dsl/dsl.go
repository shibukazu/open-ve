package dsl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis"
	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"google.golang.org/protobuf/proto"
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
		return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("Unsupported variable type: %s\nPlease specify one of the following types: int, uint, double, bool, bytes", v.Type))
	}
}

type Validation struct {
	ID        string     `yaml:"id" json:"id"`
	Cels      []string   `yaml:"cels" json:"cels"`
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
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return bytes, nil
}

func DeserializeVariables(bytes []byte) ([]Variable, error) {
	var vars []Variable
	err := json.Unmarshal(bytes, &vars)
	if err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
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
	Validations []Validation `yaml:"validations" json:"validations"`
}

type DSLReader struct {
	redis  *redis.Client
	logger *slog.Logger
}

func NewDSLReader(logger *slog.Logger, redis *redis.Client) *DSLReader {
	return &DSLReader{redis: redis, logger: logger}
}

func (r *DSLReader) Read(ctx context.Context) (*DSL, error) {
	dsl := &DSL{}
	dslJSON, err := r.redis.Get("schema").Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrDSLNotFound)
	}

	if err := json.Unmarshal(dslJSON, dsl); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return dsl, nil
}

func (r *DSLReader) Register(ctx context.Context, dsl *DSL) error {
	if err := r.resetRedis(); err != nil {
		return err
	}
	if err := r.parseAndSaveDSL(dsl); err != nil {
		return err
	}
	return nil
}

func (r *DSLReader) GetVariableNameToCELType(ctx context.Context, id string) (map[string]string, error) {
	dsl, err := r.Read(ctx)
	if err != nil {
		return nil, err
	}

	validation := Validation{}
	for _, v := range dsl.Validations {
		if v.ID == id {
			validation = v
			break
		}
	}
	variables := validation.Variables

	var variableNameToCELType = make(map[string]string)
	for _, v := range variables {
		variableNameToCELType[v.Name] = v.Type
	}
	return variableNameToCELType, nil
}

func (r *DSLReader) resetRedis() error {
	if err := r.redis.FlushDB().Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	return nil
}

func (r *DSLReader) parseAndSaveDSL(dsl *DSL) error {
	// Save DSL to Redis and Invalid HTML Escape
	var dslJson bytes.Buffer
	enc := json.NewEncoder(&dslJson)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(dsl); err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	if err := r.redis.Set("schema", dslJson.String(), 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	for _, v := range dsl.Validations {
		// Save Variables to Redis
		variablesBytes, err := v.serializeVariables()
		if err != nil {
			return err
		}
		if err := r.redis.Set(GetVariablesID(v.ID), variablesBytes, 0).Err(); err != nil {
			return failure.Translate(err, appError.ErrRedisOperationFailed)
		}

		celVariables, err := ToCELVariables(v.Variables)
		if err != nil {
			return err
		}
		env, err := cel.NewEnv(celVariables...)
		if err != nil {
			return failure.Translate(err, appError.ErrCELSyantaxError)
		}

		allASTBytes := make([][]byte, 0, len(v.Cels))
		for _, inputCel := range v.Cels {
			ast, issues := env.Compile(inputCel)
			if issues != nil && issues.Err() != nil {
				return failure.Translate(issues.Err(), appError.ErrCELSyantaxError)
			}

			// Convert AST to Proto
			expr, err := cel.AstToCheckedExpr(ast)
			if err != nil {
				return failure.Translate(err, appError.ErrCELSyantaxError)
			}
			astBytes, err := proto.Marshal(expr)
			if err != nil {
				return failure.Translate(err, appError.ErrCELSyantaxError)
			}
			allASTBytes = append(allASTBytes, astBytes)
		}

		encodedAllASTBytes, err := encodeAllASTBytes(allASTBytes)
		if err != nil {
			return err
		}

		// Save AST to Redis
		if err := r.redis.Set(GetASTID(v.ID), encodedAllASTBytes, 0).Err(); err != nil {
			return failure.Translate(err, appError.ErrRedisOperationFailed)
		}
	}

	return nil
}

func encodeAllASTBytes(allASTBytes [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(allASTBytes); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return buf.Bytes(), nil
}
