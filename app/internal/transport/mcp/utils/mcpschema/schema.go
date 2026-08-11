package mcpschema

import (
	"fmt"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
)

// typeSchemas overrides schema inference for types whose reflected shape does not
// match their JSON representation. uuid.UUID is [16]byte, so inference describes it
// as an array of bytes, while MarshalJSON writes a string: the SDK then rejects its
// own response when validating it against the generated output schema.
var typeSchemas = map[reflect.Type]*jsonschema.Schema{
	reflect.TypeFor[uuid.UUID](): {Type: "string", Format: "uuid"},
}

// OutputFor builds a tool output schema explicitly, so that AddTool skips inference
// and uses it as is. It panics on failure: schemas are built once at startup from
// static types, so an error means the types cannot be described at all.
func OutputFor[T any]() *jsonschema.Schema {
	schema, err := jsonschema.For[T](&jsonschema.ForOptions{TypeSchemas: typeSchemas})
	if err != nil {
		panic(fmt.Errorf("build output schema for %s: %w", reflect.TypeFor[T](), err))
	}

	return schema
}
