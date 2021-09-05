package context_reader

import (
	"context"
	"io"
)

func ReadAll(ctx context.Context, r io.Reader) ([]byte, error) {
	// на основе io.ReadAll

	b := make([]byte, 0, 512)
	var err error

	chEndPart := make(chan struct{})

	readPart := func() {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n := 0
		n, err = r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		chEndPart <- struct{}{}
	}

	for {
		go readPart()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-chEndPart:
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				return b, err
			}
		}
	}
}
