package unpack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	type target struct {
		Field string
	}
	type notTarget struct{}

	reg := NewRegister(func(r *Register, v any, data map[string]any) error {
		switch vv := v.(type) {
		case *target:
			vv.Field = data["field"].(string)
		default:
			return ErrNotFound
		}
		return nil
	})
	reg.AddMutator("prefix_no_args", func(v string, args MutatorArgs) (string, error) {
		return "prefixed_" + v, nil
	})
	reg.AddMutator("prefix_with_args", func(v string, args MutatorArgs) (string, error) {
		if err := args.Require("first", "final"); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s_%s_%s", args.String("first"), args.String("final"), v), nil
	})

	t.Run("NewRegister", func(t *testing.T) {
		require.IsType(t, &Register{}, reg)
	})

	t.Run("AddMutator", func(t *testing.T) {
		require.Len(t, reg.mutators, 2)
		require.Contains(t, reg.mutators, "prefix_no_args")
		require.Contains(t, reg.mutators, "prefix_with_args")
	})

	t.Run("ApplyMutator", func(t *testing.T) {
		t.Run("prefix without args", func(t *testing.T) {
			out, err := reg.ApplyMutator("prefix_no_args", "value", nil)

			require.Nil(t, err)
			require.Equal(t, "prefixed_value", out)
		})

		t.Run("prefix missing args", func(t *testing.T) {
			out, err := reg.ApplyMutator("prefix_with_args", "value", nil)
			require.ErrorContains(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: first, final")
			require.Equal(t, "", out)

			out, err = reg.ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"final": "prefix2",
			})
			require.Error(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: first")
			require.Equal(t, "", out)

			out, err = reg.ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"first": "prefix1",
			})
			require.Error(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: final")
			require.Equal(t, "", out)
		})

		t.Run("prefix with args", func(t *testing.T) {
			out, err := reg.ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"first": "prefix1",
				"final": "prefix2",
			})
			require.Nil(t, err)
			require.Equal(t, "prefix1_prefix2_value", out)
		})
	})

	t.Run("Unpack", func(t *testing.T) {
		t.Run("found", func(t *testing.T) {
			var targ target
			err := reg.Unpack(&targ, map[string]any{
				"field": "field value",
			})

			require.Nil(t, err)
			require.Equal(t, "field value", targ.Field)
		})

		t.Run("not_found", func(t *testing.T) {
			var targ notTarget
			err := reg.Unpack(&targ, map[string]any{
				"field": "field value",
			})

			require.ErrorIs(t, err, ErrNotFound)
		})
	})
}

func TestMutatorArgMap(t *testing.T) {
	args := MutatorArgMap{
		"bool":   true,
		"int":    100,
		"float":  3.14,
		"string": "this is a string",
	}

	t.Run("Bool", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			require.True(t, args.Bool("bool"))
			require.True(t, args.Bool("bool", false))
		})

		t.Run("not exists", func(t *testing.T) {
			require.False(t, args.Bool("ghost"))
			require.True(t, args.Bool("ghost", true))
		})

		t.Run("exists wrong type", func(t *testing.T) {
			require.False(t, args.Bool("string"))
			require.True(t, args.Bool("string", true))
		})
	})

	t.Run("Int", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			require.Equal(t, 100, args.Int("int"))
			require.Equal(t, 100, args.Int("int", 200))
		})

		t.Run("not exists", func(t *testing.T) {
			require.Equal(t, 0, args.Int("ghost"))
			require.Equal(t, 200, args.Int("ghost", 200))
		})

		t.Run("exists wrong type", func(t *testing.T) {
			require.Equal(t, 0, args.Int("string"))
			require.Equal(t, 200, args.Int("string", 200))
		})
	})

	t.Run("Float", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			require.Equal(t, 3.14, args.Float("float"))
			require.Equal(t, 3.14, args.Float("float", 53.2))
		})

		t.Run("not exists", func(t *testing.T) {
			require.Equal(t, 0.0, args.Float("ghost"))
			require.Equal(t, 3.14, args.Float("ghost", 3.14))
		})

		t.Run("exists wrong type", func(t *testing.T) {
			require.Equal(t, 0.0, args.Float("string"))
			require.Equal(t, 3.14, args.Float("string", 3.14))
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			require.Equal(t, "this is a string", args.String("string"))
			require.Equal(t, "this is a string", args.String("string", "fallback"))
		})

		t.Run("not exists", func(t *testing.T) {
			require.Equal(t, "", args.String("ghost"))
			require.Equal(t, "fallback", args.String("ghost", "fallback"))
		})

		t.Run("exists wrong type", func(t *testing.T) {
			require.Equal(t, "", args.String("int"))
			require.Equal(t, "fallback", args.String("int", "fallback"))
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("exists", func(t *testing.T) {
			require.Equal(t, true, args.Get("bool"))
			require.Equal(t, true, args.Get("bool", false))
			require.Equal(t, 100, args.Get("int"))
			require.Equal(t, 100, args.Get("int", -1))
			require.Equal(t, 3.14, args.Get("float"))
			require.Equal(t, 3.14, args.Get("float", 2.3))
			require.Equal(t, "this is a string", args.Get("string"))
			require.Equal(t, "this is a string", args.Get("string", "fallback"))
		})

		t.Run("not exists", func(t *testing.T) {
			require.Equal(t, nil, args.Get("ghost"))
			require.Equal(t, true, args.Get("ghost", true))
			require.Equal(t, 1, args.Get("ghost", 1))
			require.Equal(t, 1.1, args.Get("ghost", 1.1))
			require.Equal(t, "fallback", args.Get("ghost", "fallback"))
		})
	})

	t.Run("Has", func(t *testing.T) {
		t.Run("has single", func(t *testing.T) {
			require.True(t, args.Has("bool"))
			require.True(t, args.Has("int"))
			require.True(t, args.Has("float"))
			require.True(t, args.Has("string"))
		})

		t.Run("has multi", func(t *testing.T) {
			require.True(t, args.Has("bool", "int", "float", "string"))
		})

		t.Run("has some", func(t *testing.T) {
			require.False(t, args.Has("bool", "int", "float", "string", "ghost"))
		})

		t.Run("has none single", func(t *testing.T) {
			require.False(t, args.Has("ghost"))
		})

		t.Run("has none multi", func(t *testing.T) {
			require.False(t, args.Has("ghost", "pizza"))
		})
	})

	t.Run("Require", func(t *testing.T) {
		t.Run("has single", func(t *testing.T) {
			require.Nil(t, args.Require("bool"))
			require.Nil(t, args.Require("int"))
			require.Nil(t, args.Require("float"))
			require.Nil(t, args.Require("string"))
		})

		t.Run("has all", func(t *testing.T) {
			require.Nil(t, args.Require("bool", "int", "float", "string"))
		})

		t.Run("has some", func(t *testing.T) {
			require.ErrorContains(t,
				args.Require("bool", "int", "float", "string", "ghost", "tissue"),
				"unpack: mutator is missing the following args: ghost, tissue",
			)
		})

		t.Run("has none single", func(t *testing.T) {
			require.ErrorContains(t,
				args.Require("ghost"),
				"unpack: mutator is missing the following args: ghost",
			)
		})

		t.Run("has none multi", func(t *testing.T) {
			require.ErrorContains(t,
				args.Require("ghost", "cup", "knife"),
				"unpack: mutator is missing the following args: ghost, cup, knife",
			)
		})
	})

	t.Run("MustHave", func(t *testing.T) {
		t.Run("has single", func(t *testing.T) {
			args.MustHave("bool")
			args.MustHave("int")
			args.MustHave("float")
			args.MustHave("string")
		})

		t.Run("require has all", func(t *testing.T) {
			args.MustHave("bool", "int", "float", "string")
		})

		t.Run("require has some", func(t *testing.T) {
			require.PanicsWithValue(t,
				"unpack: mutator is missing the following args: ghost, tissue",
				func() {
					args.MustHave("bool", "int", "float", "string", "ghost", "tissue")
				},
			)
		})

		t.Run("require has none single", func(t *testing.T) {
			require.PanicsWithValue(t,
				"unpack: mutator is missing the following args: ghost",
				func() {
					args.MustHave("ghost")
				},
			)
		})

		t.Run("require has none multi", func(t *testing.T) {
			require.PanicsWithValue(t,
				"unpack: mutator is missing the following args: ghost, cup, knife",
				func() {
					args.MustHave("ghost", "cup", "knife")
				},
			)
		})
	})
}
