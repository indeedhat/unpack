package unpack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModLevelExports(t *testing.T) {
	type target struct {
		Field string
	}
	type notTarget struct{}

	require.Nil(t, defaultRegister.handler)

	RegisterHandler(func(r *Register, v any, data map[string]any) error {
		switch vv := v.(type) {
		case *target:
			vv.Field = data["field"].(string)
		default:
			return ErrNotFound
		}
		return nil
	})

	AddMutator("prefix_no_args", func(v string, args MutatorArgs) (string, error) {
		return "prefixed_" + v, nil
	})
	AddMutator("prefix_with_args", func(v string, args MutatorArgs) (string, error) {
		if err := args.Require("first", "final"); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s_%s_%s", args.String("first"), args.String("final"), v), nil
	})

	t.Run("RegisterHandler", func(t *testing.T) {
		require.NotNil(t, defaultRegister.handler)
	})

	t.Run("AddMutator", func(t *testing.T) {
		require.Len(t, defaultRegister.mutators, 2)
		require.Contains(t, defaultRegister.mutators, "prefix_no_args")
		require.Contains(t, defaultRegister.mutators, "prefix_with_args")
	})

	t.Run("ApplyMutator", func(t *testing.T) {
		t.Run("prefix without args", func(t *testing.T) {
			out, err := ApplyMutator("prefix_no_args", "value", nil)

			require.Nil(t, err)
			require.Equal(t, "prefixed_value", out)
		})

		t.Run("prefix missing args", func(t *testing.T) {
			out, err := ApplyMutator("prefix_with_args", "value", nil)
			require.ErrorContains(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: first, final")
			require.Equal(t, "", out)

			out, err = ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"final": "prefix2",
			})
			require.Error(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: first")
			require.Equal(t, "", out)

			out, err = ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"first": "prefix1",
			})
			require.Error(t, err, "unpack: mutator 'prefix_with_args' returned error: unpack: mutator is missing the following args: final")
			require.Equal(t, "", out)
		})

		t.Run("prefix with args", func(t *testing.T) {
			out, err := ApplyMutator("prefix_with_args", "value", MutatorArgMap{
				"first": "prefix1",
				"final": "prefix2",
			})
			require.Nil(t, err)
			require.Equal(t, "prefix1_prefix2_value", out)
		})
	})

	t.Run("Struct", func(t *testing.T) {
		t.Run("found", func(t *testing.T) {
			var targ target
			err := Struct(&targ, map[string]any{
				"field": "field value",
			})

			require.Nil(t, err)
			require.Equal(t, "field value", targ.Field)
		})

		t.Run("not_found", func(t *testing.T) {
			var targ notTarget
			err := Struct(&targ, map[string]any{
				"field": "field value",
			})

			require.ErrorIs(t, err, ErrNotFound)
		})
	})
}
