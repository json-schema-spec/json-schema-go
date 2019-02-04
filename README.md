# jsonschema [![][badge]][godoc]

This package provides an implementation of JSON Schema validation. In
particular, it does so with the following goals:

* **High performance.** Internally, this package pre-compiles schemas, and
  allocates these pre-compiled schemas in an arena to reduce memory use and
  cache locality.
* **Running untrusted schemas.** This package will never download schemas from
  the network, nor fetch them from a local filesystem. Furthermore, you can tell
  this package to abort early if it appears that a schema is defined cyclically.
* **Control over number of errors returned.** If you are only interested in
  knowing whether a schema is valid or not, you can have this package stop
  evaluation on the first error. If you're presenting errors to users, you can
  also limit the number of errors to some sensible amount.

[badge]: https://godoc.org/github.com/json-schema-spec/json-schema-go?status.svg
[godoc]: https://godoc.org/github.com/json-schema-spec/json-schema-go

## Documentation

You can find detailed documentation at:

https://godoc.org/github.com/json-schema-spec/json-schema-go

## Usage

Create a validator with `NewValidator`, and then perform validation using the
`Validate` method.

```go
import "github.com/json-schema-spec/json-schema-go"

func main() {
  // For demo purposes, this is a literal value. But this data format is the one
  // you get from the encoding/json package by default.
  schema := map[string]interface{}{
    "properties": map[string]interface{}{
      "name": map[string]interface{}{
        "type": "string",
        "minLength": 0,
      },
      "age": map[string]interface{}{
        "type": "integer",
      },
    },
  }

  instance := map[string]interface{}{
    "name": "", // note: this name is too short
    "age": "thirty seven", // note: this age is of the wrong type
  }

  validator, err := jsonschema.NewValidator([]map[string]interface{}{schema})
  if err != nil {
    // errors come from invalid schemas, or referring to non-existing schemas
    panic(err)
  }

  result, err := validator.Validate(instance)
  if err != nil {
    // errors come only from schemas going in an infinite loop on an instance
    panic(err)
  }

  fmt.Println(result.IsValid())
  // Output: false

  fmt.Println(result.Overflowed)
  // Output: false

  for _, err := range result.Errors {
    fmt.Printf(
      "validation error at %s (due to: %s)",
      err.InstancePath.String(), err.SchemaPath.String()
    )
  }

  // Output:
  //
  // Validation error at /age (due to: /properties/age/type)
  // Validation error at /name (due to: /properties/name/minLength)
}
```
