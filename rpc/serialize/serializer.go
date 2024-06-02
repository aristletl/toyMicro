package serialize

type Serializer interface {
	Code() byte

	Encode(val any) ([]byte, error)

	Decode([]byte, any) error
}
