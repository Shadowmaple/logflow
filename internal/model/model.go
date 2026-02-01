package model

import (
	"fmt"
	"time"

	"github.com/Shadowmaple/logflow/internal/utils"
)

type Event struct {
	Data map[string]any
	Time time.Time
}

func (e *Event) String() string {
	// TODO: 不忽略0值
	return fmt.Sprintf("%s: %v", utils.FormatTime(e.Time), e.Data)
}
