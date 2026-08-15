package codec

import "fmt"

type PlainEncoder struct{}

func (e *PlainEncoder) Encode(v any) ([]byte, error) {
	res := fmt.Sprintf("%v", v)
	return []byte(res), nil
}
