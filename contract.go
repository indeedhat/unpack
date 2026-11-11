package unpack

type Handler func(r *Register, v any, data map[string]any) error

type Mutator[T any] func(v T, args MutatorArgs) (T, error)

type MutatorArgs interface {
	Has(keys ...string) bool
	MustHave(keys ...string)
	Require(keys ...string) error

	Bool(key string, fallback ...bool) bool
	Float(key string, fallback ...float64) float64
	Get(key string, falback ...any) any
	Int(key string, fallback ...int) int
	String(key string, fallback ...string) string
}
