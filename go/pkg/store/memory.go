package store

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
)

type MemoryStore struct {
	memory map[string][]byte
	mu     sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	mamory := make(map[string][]byte)
	return &MemoryStore{memory: mamory}
}

func (s *MemoryStore) Reset() error {
	s.mu.Lock()
	s.memory = make(map[string][]byte)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) WriteSchema(dsl *dsl.DSL) error {
	var dslJson bytes.Buffer
	enc := json.NewEncoder(&dslJson)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(dsl); err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	s.mu.Lock()
	s.memory["schema"] = dslJson.Bytes()
	s.mu.Unlock()

	return nil
}

func (s *MemoryStore) ReadSchema() (*dsl.DSL, error) {
	dsl := &dsl.DSL{}
	s.mu.RLock()
	dslJSON, ok := s.memory["schema"]
	s.mu.RUnlock()
	if !ok {
		return nil, failure.New(appError.ErrRedisOperationFailed)
	}

	if err := json.Unmarshal(dslJSON, dsl); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return dsl, nil
}

func (s *MemoryStore) WriteVariables(id string, variables []dsl.Variable) error {
	variablesJson, err := json.Marshal(variables)
	if err != nil {
		return failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	s.mu.Lock()
	s.memory[getVariablesID(id)] = variablesJson
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) ReadVariables(id string) ([]dsl.Variable, error) {
	s.mu.RLock()
	variablesJSON, ok := s.memory[getVariablesID(id)]
	s.mu.RUnlock()
	if !ok {
		return nil, failure.New(appError.ErrRedisOperationFailed)
	}

	var variables []dsl.Variable
	if err := json.Unmarshal(variablesJSON, &variables); err != nil {
		return nil, failure.Translate(err, appError.ErrDSLSyntaxError)
	}
	return variables, nil
}

func (s *MemoryStore) WriteAllEncodedAST(id string, allEncodedAST [][]byte) error {
	if len(allEncodedAST) == 0 {
		return nil
	}
	jsonEncodedAllEncodedAST, err := jsonEncodeAllEncodedAST(allEncodedAST)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.memory[getAstID(id)] = jsonEncodedAllEncodedAST
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) ReadAllEncodedAST(id string) ([][]byte, error) {
	s.mu.RLock()
	jsonEncodedAllEncodedAST, ok := s.memory[getAstID(id)]
	s.mu.RUnlock()
	if !ok {
		return nil, failure.New(appError.ErrRedisOperationFailed)
	}
	return jsonDecodeAllEncodedAST(jsonEncodedAllEncodedAST)
}
