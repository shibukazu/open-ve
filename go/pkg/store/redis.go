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
	id          string
	redisClient *redis.Client
}

func NewRedisStore(id string, redisClient *redis.Client) *RedisStore {
	return &RedisStore{id: id, redisClient: redisClient}
}

func (s *RedisStore) Reset() error {
	keys, _ := s.redisClient.Scan(0, s.id+":*", 0).Val()
	if len(keys) == 0 {
		return nil
	}
	if err := s.redisClient.Del(keys...).Err(); err != nil {
		return failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to reset redis store"))
	}
	return nil
}

func (s *RedisStore) WriteSchema(dsl *dsl.DSL) error {
	var dslJson bytes.Buffer
	enc := json.NewEncoder(&dslJson)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(dsl); err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError, failure.Messagef("failed to encode dsl to json"))
	}
	if err := s.redisClient.Set(s.id+":schema", dslJson.String(), 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to save schema"))
	}
	return nil
}

func (s *RedisStore) ReadSchema() (*dsl.DSL, error) {
	dsl := &dsl.DSL{}
	dslJSON, err := s.redisClient.Get(s.id + ":schema").Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to get schema"))
	}

	if err := json.Unmarshal(dslJSON, dsl); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError, failure.Messagef("failed to decode dsl from json"))
	}
	return dsl, nil
}

func (s *RedisStore) WriteVariables(id string, variables []dsl.Variable) error {
	variablesJson, err := json.Marshal(variables)
	if err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError, failure.Messagef("failed to encode variables to json"))
	}
	if err := s.redisClient.Set(getVariablesID(s.id, id), variablesJson, 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to save variables"))
	}
	return nil
}

func (s *RedisStore) ReadVariables(id string) ([]dsl.Variable, error) {
	variablesJson, err := s.redisClient.Get(getVariablesID(s.id, id)).Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to get variables"))
	}

	var variables []dsl.Variable
	if err := json.Unmarshal(variablesJson, &variables); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError, failure.Messagef("failed to decode variables from json"))
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

	if err := s.redisClient.Set(getAstID(s.id, id), jsonEncodedAllEncodedAST, 0).Err(); err != nil {
		return failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to save all encoded AST"))
	}
	return nil
}

func (s *RedisStore) ReadAllEncodedAST(id string) ([][]byte, error) {
	jsonEncodedAllEncodedAST, err := s.redisClient.Get(getAstID(s.id, id)).Bytes()
	if err != nil {
		return nil, failure.Translate(err, appError.ErrStoreOperationFailed, failure.Messagef("failed to get all encoded AST"))
	}
	allEncodedAST, err := jsonDecodeAllEncodedAST(jsonEncodedAllEncodedAST)
	if err != nil {
		return nil, err
	}
	return allEncodedAST, nil
}
