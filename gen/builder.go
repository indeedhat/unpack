package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func buildFileOpen(w io.Writer, packageName, targetPath string) {
	fmt.Fprintf(w, "package unpack_%s\n", packageName)
	fmt.Fprintf(w, `import (
			target "%s"

			"github.com/indeedhat/unpack/registry"
		)
		`, targetPath)
}

func buildMapperOpen(w io.Writer, targetType string) {
	fmt.Fprintf(w, "func Unpack%s(t *target.%s, data map[string]any) error {\n", targetType, targetType)
}

func buildMapperClose(w io.Writer) {
	fmt.Fprint(w, "return nil\n}\n")
}

func buildEnsureNotNil(w io.Writer, targetType string) {
	fmt.Fprintf(w, `if t == nil {
			*t = target.%s{}
		}
		`,
		stripPtr(targetType),
	)
}

func buildLeteralAssignement(w io.Writer, key, fieldName, dataType string) {
	if isPtr(dataType) {
		fmt.Fprintf(w, `if v, found := data["%s"]; found {
			t.%s = new(v.(%s))
		}
		`,
			key,
			fieldName,
			stripPtr(dataType),
		)
		return
	}

	fmt.Fprintf(w, `if v, found := data["%s"]; found {
			t.%s = v.(%s)
		}
		`,
		key,
		fieldName,
		dataType,
	)
}

func buildNestedAssgnment(w io.Writer, key, fieldName, dataType string) {
	if isPtr(dataType) {
		fmt.Fprintf(w, `if v, found := data["%s"]; found {
				vv, ok := v.(map[string]any)
				if ok {
					if err := Unpack%s(t.%s, vv); err != nil {
						return err
					}
				}
			}
			`,
			key,
			dataType,
			fieldName,
		)
		return
	}

	fmt.Fprintf(w, `if v, found := data["%s"]; found {
			vv, ok := v.(map[string]any)
			if ok {
				if err := Unpack%s(&t.%s, vv); err != nil {
					return err
				}
			}
		}
		`,
		key,
		dataType,
		fieldName,
	)
}

func buildMutatorCall(w io.Writer, dataType string) {
	fmt.Fprintf(w, `if err := apply%sMutations(data); err != nil {
		return err
		}
		`,
		dataType,
	)
}

func buildMutatorOpen(w io.Writer, dataType string) {
	fmt.Fprintf(w, "func apply%sMutations(data map[string]any) error {", dataType)
}

func buildMutatorClose(w io.Writer) {
	fmt.Fprint(w, "return nil\n}\n")
}

func buildMutatorApply(w io.Writer, dataKey, typ string, mu mutator) {
	if len(mu.Args) == 0 {
		fmt.Fprintf(w, `if v, err := registry.Apply("%s", data["%s"].(%s), nil); err != nil {
				return err
			} else {
				data["%s"] = v
			}
			`,
			mu.Name,
			dataKey,
			typ,
			dataKey,
		)
		return
	}

	var mapBuf bytes.Buffer

	escape := func(v string) string {
		return strings.ReplaceAll(strings.ReplaceAll(v, `\`, `\\`), "'", `\'`)
	}

	for k, v := range mu.Args {
		fmt.Fprintf(&mapBuf, `"%s": "%s",`, escape(k), escape(v))
	}

	fmt.Fprintf(w, `if v, err := registry.Apply("%s", data["%s"].(%s), map[string]string{
		%s
		}); err != nil {
			return err
		} else {
			data["%s"] = v
		}
		`,
		mu.Name,
		dataKey,
		typ,
		mapBuf.String(),
		dataKey,
	)
}
