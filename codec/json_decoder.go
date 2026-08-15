package codec

import (
	"encoding/json"
	"errors"

	"github.com/Shadowmaple/logflow/internal/logger"

	"go.uber.org/zap"
)

type JsonDecoder struct{}

func (d *JsonDecoder) Decode(data []byte) (map[string]any, error) {
	v := make(map[string]any)
	if err := json.Unmarshal(data, &v); err != nil {
		logger.Error("JsonDecoder Decode err:"+err.Error(), zap.String("data", string(data)))
		v["message"] = string(data)
		v["@errorTag"] = "JSODecodeFailed"
		return v, errors.New("JSON decode failed")
	}
	return v, nil
}
