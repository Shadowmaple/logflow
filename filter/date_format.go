package filter

import (
	"errors"
	"fmt"
	"time"

	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"

	"go.uber.org/zap"
)

// DateFormatConfig defines the configuration structure for dateFormat filter
type DateFormatConfig struct {
	Source       string `json:"source"`
	Target       string `json:"target"`
	Format       string `json:"format"`
	Location     string `json:"location"`
	SetIfNil     string `json:"set_if_nil"`
	SetIfFail    string `json:"set_if_fail"`
	RemoveIfFail bool   `json:"remove_if_fail"`
	Overwrite    bool   `json:"overwrite"`
}

type DateFormatFilter struct {
	source       string
	target       string
	format       string
	location     *time.Location
	setIfNil     string
	setIfFail    string
	removeIfFail bool
	overwrite    bool
}

func init() {
	register("dateFormat", newDateFormatFilter)
}

func newDateFormatFilter(config map[string]any) model.Filter {
	f := &DateFormatFilter{
		overwrite: true,
	}

	// Parse configuration using SafeDecodeConfig
	var dateFormatConfig DateFormatConfig
	dateFormatConfig.Overwrite = true

	utils.SafeDecodeConfig("dateFormat", config, &dateFormatConfig)

	// Validate required fields
	if dateFormatConfig.Source == "" {
		panic("dateFormat filter: 'source' is required")
	}
	if dateFormatConfig.Target == "" {
		panic("dateFormat filter: 'target' is required")
	}
	if dateFormatConfig.Format == "" {
		panic("dateFormat filter: 'format' is required")
	}

	f.source = dateFormatConfig.Source
	f.target = dateFormatConfig.Target
	f.setIfNil = dateFormatConfig.SetIfNil
	f.setIfFail = dateFormatConfig.SetIfFail
	f.removeIfFail = dateFormatConfig.RemoveIfFail
	f.overwrite = dateFormatConfig.Overwrite

	// Convert common format (yyyy-MM-dd) to Go-style format (2006-01-02)
	f.format = utils.ConvertToGoFormat(dateFormatConfig.Format)
	if f.format == "" {
		panic("dateFormat filter: invalid format")
	}

	// Parse location
	if dateFormatConfig.Location != "" {
		var err error
		f.location, err = time.LoadLocation(dateFormatConfig.Location)
		if err != nil {
			panic(fmt.Sprintf("dateFormat filter: load location error: %s", err))
		}
	}
	return f
}

func (f *DateFormatFilter) Filter(event *event.Event) (*event.Event, error) {
	// Priority 1: set_if_nil - if source field does not exist
	sourceVal, sourceExists := event.Data[f.source]
	if !sourceExists {
		if f.setIfNil != "" {
			f.setTargetValue(event, f.setIfNil)
		}
		return event, errors.New("dateFormat filter failed: source field not found")
	}

	// Source exists, try to get time.Time value
	t, ok := f.extractTime(sourceVal)
	if !ok {
		f.handleFailure(event)
		return event, errors.New("dateFormat filter failed: source field is not of type time.Time or *time.Time")
	}

	// Apply location if configured
	if f.location != nil {
		t = t.In(f.location)
	}

	// Format the time
	formatted := t.Format(f.format)

	// Write to target
	f.setTargetValue(event, formatted)

	return event, nil
}

// extractTime tries to extract a time.Time from the given value
func (f *DateFormatFilter) extractTime(val any) (time.Time, bool) {
	var zeroTime time.Time
	if val == nil {
		return zeroTime, false
	}

	// Direct time.Time
	if t, ok := val.(time.Time); ok {
		return t, true
	}

	// Pointer to time.Time
	if tp, ok := val.(*time.Time); ok && tp != nil {
		return *tp, true
	}

	logger.Error("dateFormat filter: source field is not of type time.Time",
		zap.String("source", f.source),
		zap.String("actual_type", fmt.Sprintf("%T", val)))
	return zeroTime, false
}

// handleFailure processes the failure case according to priority:
// 1. remove_if_fail (highest priority on failure) - removes target field
// 2. set_if_fail (lower priority) - sets target to fallback value
func (f *DateFormatFilter) handleFailure(event *event.Event) {
	if f.removeIfFail {
		// Highest priority: remove target field if it exists
		delete(event.Data, f.target)
		return
	}
	if f.setIfFail != "" {
		f.setTargetValue(event, f.setIfFail)
	}
}

// setTargetValue writes value to target field respecting overwrite setting
func (f *DateFormatFilter) setTargetValue(event *event.Event, value string) {
	if !f.overwrite {
		if _, exists := event.Data[f.target]; exists {
			return
		}
	}
	event.Data[f.target] = value
}
