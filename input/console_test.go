package input

import "testing"

func TestConsoleConfigParsing(t *testing.T) {
	tests := []struct {
		name   string
		config map[any]any
		panics bool
	}{
		{
			name: "valid console config",
			config: map[any]any{
				"codec": "json",
			},
			panics: false,
		},
		{
			name:   "empty config should work",
			config: map[any]any{},
			panics: false,
		},
		{
			name: "invalid JSON structure should panic",
			config: map[any]any{
				"codec": "xxx",
			},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("newConsoleInput() should have panicked")
					}
				}()
				newConsoleInput(tt.config)
			} else {
				input := newConsoleInput(tt.config)
				consoleInput, ok := input.(*ConsoleInput)
				if !ok {
					t.Errorf("newConsoleInput() should return *ConsoleInput")
					return
				}

				if consoleInput == nil {
					t.Errorf("ConsoleInput should not be nil")
				}
			}
		})
	}
}
