package interpreter

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type interpreterTester[TInput, TExpected any] interface {
	Interpret(TInput) (TExpected, error)
}

type interpreterTestCase[TInput, TExpected any] struct {
	name     string
	input    TInput
	expected TExpected
}

func testInterpret[TInput, TExpected any](t *testing.T, cases []interpreterTestCase[TInput, TExpected], newTester func(log *slog.Logger) interpreterTester[TInput, TExpected]) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tester := newTester(slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})))
			actual, err := tester.Interpret(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(actual, tc.expected) {
				t.Fatalf("actual result and expected result differ: %s", cmp.Diff(actual, tc.expected))
			}
		})
	}
}

func pointerTo[T any](v T) *T {
	return &v
}
