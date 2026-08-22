// Задача прочитать список url

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

func main() {
	ctx := context.Background()

	var urls = []string{
		"http://ozon.ru",
		"https://ozon.ru",
		"http://google.com",
		"http://fake.domain.tld",
		"http://ya.ru",
		"https://ya.ru",
		"http://ёёёёё",
	}

	channelSize := 3
	urlInput := Generator(ctx, urls, channelSize)

	workersCount := 2
	resChan := Start(ctx, workersCount, urlInput, func(currentUrl string) result {
		resp, err := http.Get(currentUrl)

		if err != nil {
			return result{
				url: currentUrl,
				err: fmt.Errorf("failed %s, error - %v", currentUrl, err),
			}
		}

		if resp.StatusCode != http.StatusOK {
			return result{
				url: currentUrl,
				err: fmt.Errorf("failed %s with http code %d", currentUrl, resp.StatusCode),
			}
		}

		return result{
			url: currentUrl,
		}
	})

	for r := range resChan {
		fmt.Printf("URL: %s, Error: %v\n", r.url, r.err)
	}

}

type result struct {
	url string
	err error
}

// Превращаем слайс в канал
func Generator(ctx context.Context, data []string, size int) <-chan string {
	res := make(chan string, size)

	go func() {
		defer close(res)

		for i := 0; i < len(data); i++ {
			select {
			case res <- data[i]:
			case <-ctx.Done():
				return
			}
		}
	}()

	return res
}

// Worker pool
func Start(ctx context.Context, workersCount int, input <-chan string, transform func(e string) result) <-chan result {
	res := make(chan result)

	wg := new(sync.WaitGroup)
	wg.Add(workersCount)

	for i := 0; i < workersCount; i++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-input:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					case res <- transform(v):
					}
				}
			}
		}()
	}

	go func() {
		defer close(res)
		wg.Wait()
	}()

	return res
}
