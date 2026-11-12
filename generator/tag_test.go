package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagParser(t *testing.T) {
	t.Run("missing tag", func(t *testing.T) {
		parser := newTagParser("unpack")
		tag := parser.Parse("`not_unpack:\"true\"`")

		require.Equal(t, "", tag.Key)
		require.Len(t, tag.Mutators, 0)
		require.False(t, tag.HasDive())
	})

	t.Run("empty tag", func(t *testing.T) {
		parser := newTagParser("unpack")
		tag := parser.Parse("`unpack:\"\"`")

		require.Equal(t, "", tag.Key)
		require.Len(t, tag.Mutators, 0)
		require.False(t, tag.HasDive())
	})

	t.Run("field name only", func(t *testing.T) {
		parser := newTagParser("unpack")
		tag := parser.Parse("`unpack:\"field_name\"`")

		require.Equal(t, "field_name", tag.Key)
		require.Len(t, tag.Mutators, 0)
		require.False(t, tag.HasDive())
	})

	t.Run("with mutators", func(t *testing.T) {
		parser := newTagParser("unpack")
		tag := parser.Parse("`unpack:\"field_name;trim;clamp,min=1,max=2\"`")

		require.Equal(t, "field_name", tag.Key)
		require.Len(t, tag.Mutators, 2)
		require.False(t, tag.HasDive())

		muTrim := tag.Mutators[0]
		require.Equal(t, "trim", muTrim.Name)
		require.Len(t, muTrim.Args, 0)

		muClamp := tag.Mutators[1]
		require.Equal(t, "clamp", muClamp.Name)
		require.Len(t, muClamp.Args, 2)
		require.Contains(t, muClamp.Args, "min")
		require.Contains(t, muClamp.Args, "max")
		require.Equal(t, "1", muClamp.Args["min"])
		require.Equal(t, "2", muClamp.Args["max"])
	})

	t.Run("with dive", func(t *testing.T) {
		parser := newTagParser("unpack")
		tag := parser.Parse("`unpack:\"field_name;dive\"`")

		require.Equal(t, "field_name", tag.Key)
		require.Len(t, tag.Mutators, 1)
		require.True(t, tag.HasDive())
	})
}
