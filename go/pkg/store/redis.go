package store

import (
	"bytes"
	"encoding/json"

	"github.com/go-redis/redis"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
)

type RedisStore struct {
	redisClient *redis.Client
}

func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{redisClient: redisClient}
}

func (s *RedisStore) Reset() error {
	if err := s.redisClient.FlushDB().Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	return nil
}

func (s *RedisStore) WriteSchema(dsl *dsl.DSL) error {
	var dslJson bytes.Buffer
	enc := json.NewEncoder(&dslJson)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(dsl); err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	if err := s.redisClient.Set("schema", dslJson.String(), 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	return nil
}

func (s *RedisStore) ReadSchema() (*dsl.DSL, error) {
	dsl := &dsl.DSL{}
	dslJSON, err := s.redisClient.Get("schema").Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrRedisOperationFailed)
	}

	if err := json.Unmarshal(dslJSON, dsl); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return dsl, nil
}

func (s *RedisStore) WriteVariables(id string, variables []dsl.Variable) error {
	variablesJson, err := json.Marshal(variables)
	if err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	if err := s.redisClient.Set(getVariablesID(id), variablesJson, 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	return nil
}

func (s *RedisStore) ReadVariables(id string) ([]dsl.Variable, error) {
	variablesJson, err := s.redisClient.Get(getVariablesID(id)).Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrRedisOperationFailed)
	}

	var variables []dsl.Variable
	if err := json.Unmarshal(variablesJson, &variables); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return variables, nil
}

func (s *RedisStore) WriteAllEncodedAST(id string, allEncodedAST [][]byte) error {
	if len(allEncodedAST) == 0 {
		return nil
	}

	jsonEncodedAllEncodedAST, err := jsonEncodeAllEncodedAST(allEncodedAST)
	if err != nil {
		return err
	}

	if err := s.redisClient.Set(getAstID(id), jsonEncodedAllEncodedAST, 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	return nil
}

func (s *RedisStore) ReadAllEncodedAST(id string) ([][]byte, error) {
	jsonEncodedAllEncodedAST, err := s.redisClient.Get(getAstID(id)).Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrRedisOperationFailed)
	}
	allEncodedAST, err := jsonDecodeAllEncodedAST(jsonEncodedAllEncodedAST)
	if err != nil {
		return nil, err
	}
	return allEncodedAST, nil
}


