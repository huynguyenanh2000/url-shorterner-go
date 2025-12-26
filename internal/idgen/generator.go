package idgen

type Client interface {
	Generate() uint64
}
