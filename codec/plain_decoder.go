package codec

type PlainDecoder struct{}

func (d *PlainDecoder) Decode(data []byte) (map[string]any, error) {
	return map[string]any{
		"message": string(data),
	}, nil
}
