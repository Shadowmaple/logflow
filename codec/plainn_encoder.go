package codec

type PlainEncoder struct{}

func (e *PlainEncoder) Encode(v any) ([]byte, error) {
	return []byte(v.(string)), nil
}
