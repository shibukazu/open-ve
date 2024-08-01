package store

import "github.com/shibukazu/open-ve/go/pkg/dsl"

type Store interface {
	Reset() error
	WriteSchema(*dsl.DSL) error
	ReadSchema() (*dsl.DSL, error)
	WriteVariables(string, []dsl.Variable) error
	ReadVariables(string) ([]dsl.Variable, error)
	WriteAllEncodedAST(string, [][]byte) error
	ReadAllEncodedAST(string) ([][]byte, error)
}
