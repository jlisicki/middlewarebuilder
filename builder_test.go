package middlewarebuilder

import (
	"errors"
	"testing"
)

type (
	textCreator interface {
		CreateText(input string) string
	}
	exampleMiddleware struct {
		ExtraText string
		Next      textCreator
	}
	exampleMiddlewareFactory struct {
		ExtraText string
	}
	exampleHandler struct{}
)

var errExample = errors.New("example error")

func (e exampleMiddlewareFactory) Create(next textCreator) (textCreator, error) {
	return exampleMiddleware{Next: next, ExtraText: e.ExtraText}, nil
}

func (e exampleHandler) CreateText(input string) string {
	return input + ": handler"
}

func (e exampleMiddleware) CreateText(input string) string {
	input = input + ": " + e.ExtraText
	return e.Next.CreateText(input)
}

func TestBuilder_Build(t *testing.T) {
	t.Run("Should return error when handler is not set", func(t *testing.T) {
		_, err := (&Builder[textCreator]{}).Build()
		if err == nil {
			t.Error("Expected error about missing handler but got nil")
		}
	})
	t.Run("Should return error from middlewarebuilder factory", func(t *testing.T) {
		b := &Builder[textCreator]{}
		b.
			Add(exampleMiddlewareFactory{ExtraText: "first"}).
			Add(FactoryFunc[textCreator](func(next textCreator) (textCreator, error) {
				return nil, errExample
			})).
			Add(exampleMiddlewareFactory{ExtraText: "third"}).
			WithHandler(exampleHandler{})
		_, err := b.Build()
		if !errors.Is(err, errExample) {
			t.Errorf("Expected example error but got: %v", err)
		}
	})
	t.Run("Should create middlewarebuilder chain in order", func(t *testing.T) {
		b := &Builder[textCreator]{}
		b.
			Add(exampleMiddlewareFactory{ExtraText: "first"}).
			Add(exampleMiddlewareFactory{ExtraText: "second"}).
			Add(exampleMiddlewareFactory{ExtraText: "third"}).
			WithHandler(exampleHandler{})
		chain, err := b.Build()
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		out := chain.CreateText("input")
		expected := "input: first: second: third: handler"
		if out != expected {
			t.Errorf("Got '%s' but expected '%s'", out, expected)
		}
	})
}
