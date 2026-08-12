package codec

type Decoder interface {
	Decode([]byte) (map[string]any, error)
}

func NewDecoder(codeType string) Decoder {
	switch codeType {
	case "json":
		return &JsonDecoder{}
	case "plain":
		return &PlainDecoder{}
	}
	panic("invalid codec type: " + codeType)
}
