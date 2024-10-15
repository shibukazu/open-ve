package generator

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
)

// Support Only OpenAPI 2.0
func GenerateFromOpenAPI2(filePath string) (*dsl.DSL, error) {
	validations := make([]dsl.Validation, 0)

	swaggerDoc, err := loads.Spec(filePath)
	if err != nil {
		return nil, err
	}

	paths := swaggerDoc.Spec().Paths
	for path, pathItem := range paths.Paths {
		log.Printf("Parsing Path: %s\n", path)
		if pathItem.Post != nil {
			for _, param := range pathItem.Post.Parameters {
				if param.Schema != nil {
					schema, refName, err := resolveSchemaReference(swaggerDoc.Spec(), param.Schema)
					if err != nil {
						return nil, err
					}
					variables := make([]dsl.Variable, 0)
					parseParamSchema(swaggerDoc.Spec(), schema, refName, "", &variables)
					validation := dsl.Validation{
						ID:        refName,
						Cels:      []string{},
						Variables: variables,
					}
					validations = append(validations, validation)
				}
			}
		}
	}
	return &dsl.DSL{
		Validations: validations,
	}, nil
}

func resolveSchemaReference(doc *spec.Swagger, schema *spec.Schema) (*spec.Schema, string, error) {
	if schema.Ref.String() != "" {
		ref, err := spec.ResolveRef(doc, &schema.Ref)
		if err != nil {
			return nil, "", err
		}

		refParts := strings.Split(schema.Ref.String(), "/")
		if len(refParts) > 0 {
			objectName := refParts[len(refParts)-1]
			return ref, objectName, nil
		}
		return ref, "", nil
	}
	return schema, "", nil
}

func parseParamSchema(doc *spec.Swagger, schema *spec.Schema, parentObjectName string, propName string, variables *[]dsl.Variable) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if schema.Properties != nil {
		// Object
		for propName, prop := range schema.Properties {
			if prop.Ref.String() != "" {
				resolvedProp, objectName, err := resolveSchemaReference(doc, &prop)
				if err != nil {
					return err
				}
				parentObjectName := objectName
				parseParamSchema(doc, resolvedProp, parentObjectName, "", variables)
			} else {
				if prop.Properties != nil {
					// Object
					parentObjectName := propName
					parseParamSchema(doc, &prop, parentObjectName, "", variables)
				} else {
					// Primitive
					parseParamSchema(doc, &prop, parentObjectName, propName, variables)
				}
			}
		}

	} else if schema.Items != nil {
		// TODO: Support Array
		return fmt.Errorf("OpenAPI Array Schema is not supported yet")
	} else {
		// Primitive
		if schema.Type != nil {
			if len(schema.Type) != 1 {
				return fmt.Errorf("schema.Type is not 1")
			}
			typeName := schema.Type[0]
			variable := dsl.Variable{
				Name: parentObjectName + "." + propName,
				Type: typeName,
			}
			*variables = append(*variables, variable)
		}
	}
	return nil
}