package codec

type Encoder interface {
	Encode(v any) ([]byte, error)
}

func NewEncoder(codeType string) Encoder {
	switch codeType {
	case "json":
		return &JsonEncoder{}
	case "plain":
		return &PlainEncoder{}
	}
	panic("invalid codec type: " + codeType)
}
