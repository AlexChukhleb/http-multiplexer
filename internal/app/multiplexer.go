package app

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/ach/http-multiplexer/pkg/context_reader"
	"github.com/ach/http-multiplexer/pkg/semaphore"
)

func (a *App) DoRequests(ctx context.Context, list []string) (map[string]string, error) {
	//должно быть map[string][]byte, но в json ответе значения будут в base64

	if len(list) > a.config.MaxLinksPerRequest {
		return nil, errors.New("number of requests exceeded")
	}

	sem := semaphore.NewSemaphore(a.config.MaxOutgoingConnections)

	client := &http.Client{Transport: &http.Transport{}}

	outMutex := sync.Mutex{}
	out := make(map[string]string, len(list))

	chOk := make(chan struct{}, len(list))
	chErr := make(chan error)

	for _, url := range list {
		url := url
		go func() {
			// ограничиваем количество одновременных запросов
			err := sem.Acquire(ctx, a.config.RequestTimeout*time.Duration(a.config.MaxLinksPerRequest))
			if err != nil {
				chErr <- err
				return
			}
			defer sem.Release()

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				chErr <- err
				return
			}
			res, err := client.Do(req)
			if err != nil {
				chErr <- err
				return
			}

			ctx, cancelFunc := context.WithTimeout(ctx, a.config.RequestTimeout)
			defer func() {
				cancelFunc()
			}()
			body, err := context_reader.ReadAll(ctx, res.Body)
			if err != nil {
				chErr <- err
				return
			}

			{
				outMutex.Lock()
				out[url] = string(body)
				outMutex.Unlock()
			}

			chOk <- struct{}{}
		}()
	}

	resCount := len(list)
	for {
		select {
		case <-chOk:
			resCount-- // тут можно использовать sync.WaitGroup, но мне так нравится больше
			if resCount == 0 {
				return out, nil
			}
		case err := <-chErr:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
