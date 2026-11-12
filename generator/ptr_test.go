package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPtr(t *testing.T) {
	t.Run("is pointer", func(t *testing.T) {
		require.True(t, isPtr("*string"))
	})

	t.Run("not pointer", func(t *testing.T) {
		require.False(t, isPtr("string"))
	})
}

func TestStripPtr(t *testing.T) {
	t.Run("is pointer", func(t *testing.T) {
		require.Equal(t, "string", stripPtr("*string"))
	})

	t.Run("not pointer", func(t *testing.T) {
		require.Equal(t, "string", stripPtr("string"))
	})
}
