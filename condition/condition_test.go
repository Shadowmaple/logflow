package condition

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestConditionFilterCheck(t *testing.T) {
	event := &event.Event{Data: map[string]any{
		"level": "DEBUG",
		"msg":   "hello",
		"tag":   "backup",
		"items": []any{"test", 1},
	}}

	filter := NewConditionFilter(map[any]any{
		"if": []any{
			"EQ(level, \"DEBUG\")",
			"EXIST(msg)",
			"EQ(tag, \"backup\") || EQ(tag, \"prod\")",
			"In(items, \"test\")",
		},
	})

	if filter == nil {
		t.Fatal("expected condition filter to be created")
	}

	if !filter.Check(event) {
		t.Fatal("expected all conditions to pass")
	}
}

func TestConditionFilterCheckWithAndAndNotIn(t *testing.T) {
	event := &event.Event{Data: map[string]any{
		"level": "INFO",
		"items": []any{"test", 2},
	}}

	filter := NewConditionFilter(map[any]any{
		"if": []any{
			"EQ(level, \"INFO\")",
			"NotIn(items, \"prod\")",
		},
	})

	if !filter.Check(event) {
		t.Fatal("expected condition filter to pass")
	}

	filter = NewConditionFilter(map[any]any{
		"if": []any{
			"EQ(level, \"INFO\")",
			"NotIn(items, \"test\")",
		},
	})

	if filter.Check(event) {
		t.Fatal("expected condition filter to fail when NotIn matches")
	}
}

func TestConditionFilterCheckWithNil(t *testing.T) {
	event := &event.Event{Data: map[string]any{
		"value": nil,
	}}

	filter := NewConditionFilter(map[any]any{
		"if": []any{
			"EQ(value, nil)",
		},
	})

	if !filter.Check(event) {
		t.Fatal("expected nil equality to pass")
	}
}

func TestConditionFilterCheckWithMixedAndOr(t *testing.T) {
	event := &event.Event{Data: map[string]any{
		"level": "INFO",
		"msg":   "hello",
	}}

	filter := NewConditionFilter(map[any]any{
		"if": []any{
			"EQ(level, \"DEBUG\") || EQ(level, \"INFO\")",
			"EXIST(msg)",
		},
	})

	if !filter.Check(event) {
		t.Fatal("expected mixed OR and AND semantics to pass")
	}
}

func TestConditionBuildRPN(t *testing.T) {
	raws := []string{
		`EXIST(msg)`,
		`EXIST(msg) && IN(items, "test")`,
		`EQ(tag, "backup") || EQ(level, 123)`,
		`EQ(level, "DEBUG") || EXIST(msg) && EQ(tag, "backup")`,
	}

	expectedRPNs := [][]any{
		{"EXIST(msg)"},
		{"EXIST(msg)", "IN(items, \"test\")", _op_and},
		{"EQ(tag, \"backup\")", "EQ(level, 123)", _op_or},
		{"EQ(level, \"DEBUG\")", "EXIST(msg)", "EQ(tag, \"backup\")", _op_and, _op_or},
	}

	for i, raw := range raws {
		rpn, err := buildRPN(raw)
		if err != nil {
			t.Fatalf("unexpected error for raw %q: %v", raw, err)
		}
		if len(rpn) != len(expectedRPNs[i]) {
			t.Fatalf("expected RPN length %d, got %d for raw %q", len(expectedRPNs[i]), len(rpn), raw)
		}
		for j, token := range rpn {
			switch token := token.(type) {
			case string:
				if token != expectedRPNs[i][j] {
					t.Fatalf("expected RPN token %v at position %d, got %v for raw %q", expectedRPNs[i][j], j, token, raw)
				}
			case uint8:
				exp := uint8(expectedRPNs[i][j].(int))
				if token != exp {
					t.Fatalf("expected RPN operator %v at position %d, got %v for raw %q", exp, j, token, raw)
				}
			case int:
				exp := expectedRPNs[i][j].(int)
				if token != exp {
					t.Fatalf("expected RPN operator %v at position %d, got %v for raw %q", exp, j, token, raw)
				}
			default:
				t.Fatalf("unexpected RPN token type %T for raw %q", token, raw)
			}
		}
	}
}

func TestParseConditionExprTreeWithAndOr(t *testing.T) {
	expr, err := parseToConditionNode(`EXIST(msg) && EQ(level, 12) || EQ(tag, "backup")`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if expr == nil {
		t.Fatal("expected parsed expression tree to be non-nil")
	}
	if expr.op != _op_or {
		t.Fatalf("expected root operator %d, got %d", _op_or, expr.op)
	}
	if expr.left == nil || expr.right == nil {
		t.Fatal("expected both left and right branches to be present")
	}
	if expr.left.op != _op_and {
		t.Fatalf("expected left branch to be an AND expression, got %d", expr.left.op)
	}
	if expr.left.left == nil || expr.left.right == nil {
		t.Fatal("expected AND branch to have both operands")
	}
	if expr.right.clause == nil {
		t.Fatal("expected right branch to be a clause")
	}

	cases := []struct {
		name string
		data map[string]any
		want bool
	}{
		{
			name: "matches AND branch",
			data: map[string]any{"msg": "hello", "level": 12},
			want: true,
		},
		{
			name: "matches OR branch",
			data: map[string]any{"tag": "backup"},
			want: true,
		},
		{
			name: "does not match when both branches fail",
			data: map[string]any{"msg": "hello", "level": "DEBUG"},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := expr.check(tc.data); got != tc.want {
				t.Fatalf("expected match %v for data %v, got %v", tc.want, tc.data, got)
			}
		})
	}
}

func TestParseConditionExprTreeWithSingle(t *testing.T) {
	expr, err := parseToConditionNode(`EXIST(msg)`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if expr == nil {
		t.Fatal("expected parsed expression tree to be non-nil")
	}
	if expr.clause == nil {
		t.Fatal("expected to be a clause")
	}

	cases := []struct {
		name string
		data map[string]any
		want bool
	}{
		{
			name: "matches AND branch",
			data: map[string]any{"msg": "hello", "level": 12},
			want: true,
		},
		{
			name: "matches OR branch",
			data: map[string]any{"tag": "backup"},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := expr.check(tc.data); got != tc.want {
				t.Fatalf("expected match %v for data %v, got %v", tc.want, tc.data, got)
			}
		})
	}
}
