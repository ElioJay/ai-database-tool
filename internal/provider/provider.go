package provider

import "context"

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
	var out string
	for chunk := range ch {
		if chunk.Err != nil {
			return out, chunk.Err
		}
		if chunk.Done {
			break
		}
		if chunk.Delta != "" {
			out += chunk.Delta
			if render != nil {
				render(chunk.Delta)
			}
		}
	}
	return out, nil
}
