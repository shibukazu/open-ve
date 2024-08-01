package reader

import (
	"context"
	"log/slog"

	"github.com/google/cel-go/cel"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	dslPkg "github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/store"
	"google.golang.org/protobuf/proto"
)

type DSLReader struct {
	store  store.Store
	logger *slog.Logger
}

func NewDSLReader(logger *slog.Logger, store store.Store) *DSLReader {
	return &DSLReader{logger: logger, store: store}
}

func (r *DSLReader) Read(ctx context.Context) (*dslPkg.DSL, error) {
	dsl, err := r.store.ReadSchema()
	if err != nil {
		return nil, err
	}
	return dsl, nil
}

func (r *DSLReader) Register(ctx context.Context, dsl *dslPkg.DSL) error {
	if err := r.store.Reset(); err != nil {
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

	validation := dslPkg.Validation{}
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

func (r *DSLReader) parseAndSaveDSL(dsl *dslPkg.DSL) error {
	// TODO: make following operations atomic
	// Save DSL to Store
	if err := r.store.WriteSchema(dsl); err != nil {
		return err
	}
	for _, v := range dsl.Validations {
		// Save Variables to Store
		if err := r.store.WriteVariables(v.ID, v.Variables); err != nil {
			return err
		}

		celVariables, err := dslPkg.ToCELVariables(v.Variables)
		if err != nil {
			return err
		}
		env, err := cel.NewEnv(celVariables...)
		if err != nil {
			return failure.Translate(err, appError.ErrCELSyantaxError)
		}

		allEncodedAST := make([][]byte, 0, len(v.Cels))
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
			encodedAST, err := proto.Marshal(expr)
			if err != nil {
				return failure.Translate(err, appError.ErrCELSyantaxError)
			}
			allEncodedAST = append(allEncodedAST, encodedAST)
		}

		// Save All Encoded AST to Store
		if err := r.store.WriteAllEncodedAST(v.ID, allEncodedAST); err != nil {
			return err
		}
	}

	return nil
}
