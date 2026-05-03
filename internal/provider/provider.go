package provider

import (
	"context"
	"strings"
)

type Message struct {
	Role    string
	Content string
}

type Chunk struct {
	Delta string
	Done  bool
	Err   error
}

type Provider interface {
	Name() string
	Stream(ctx context.Context, msgs []Message) (<-chan Chunk, error)
}

func Collect(ch <-chan Chunk, render func(string)) (string, error) {
	var b strings.Builder
	for chunk := range ch {
		if chunk.Err != nil {
			return b.String(), chunk.Err
		}
		if chunk.Done {
			break
		}
		if chunk.Delta != "" {
			b.WriteString(chunk.Delta)
			if render != nil {
				render(chunk.Delta)
			}
		}
	}
	return b.String(), nil
}
