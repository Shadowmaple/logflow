package event

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Data map[string]any
	Time time.Time
}

func (e *Event) String() string {
	// TODO: 不忽略0值
	jsonStr, err := json.Marshal(e.Data)
	if err != nil {
		return fmt.Sprintf("%v", e.Data)
	}
	return fmt.Sprintf("%v", string(jsonStr))
}
