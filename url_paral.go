package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Задача - распараллелить запросы
// func main() {
// 	_, _ = http.Get("https://google.com")
// 	_, _ = http.Get("https://ya.ru")
// }

// Настройка http клиента
// Все эти настройки управляют пулом TCP-соединений (Connection Pool) в Go-клиенте.
// Они нужны для того, чтобы не создавать новое сетевое соединение на каждый HTTP-запрос, а переиспользовать уже открытые сокеты.
// Это экономит процессорное время, сетевой трафик и защищает операционную систему от исчерпания портов.
var customHTTPClient = &http.Client{
	Timeout: 5 * time.Second, // Глобальный таймаут на весь запрос + чтение тела
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,  // Время на установку TCP-соединения
			KeepAlive: 30 * time.Second, // Задает интервал времени, с которым операционная система отправляет пустые «проверочные» пакеты (ping) во внешнюю сеть по открытому сокету
		}).DialContext,
		MaxIdleConns:          100,              // Пул переиспользуемых соединений. Это глобальный лимит на кэш соединений, чтобы ваш клиент не держал открытыми тысячи сокетов вхолостую, расходуя оперативную память сервера.
		MaxIdleConnsPerHost:   100,              // Максимальное количество простаивающих соединений для каждого отдельного хоста.  В стандартном клиенте Go это значение по умолчанию равно 2. Если ваш микросервис отправляет к одному API 50 одновременных запросов, Go откроет 50 соединений. Но когда они завершатся, клиент сохранит в пул только 2, а остальные 48 соединений жестко закроет. На следующие 50 запросов ему придется заново устанавливать TCP- и TLS-соединения. Для нагруженных систем это значение всегда нужно увеличивать.
		IdleConnTimeout:       90 * time.Second, // Время, в течение которого неиспользуемое (простаивающее) соединение может лежать в пуле, прежде чем клиент сам решит его закрыть. Защита от «протухания» пула. Если вы сделали запрос к какому-то сайту, и больше к нему не обращаетесь, нет смысла вечно держать сокет открытым. По истечении этого таймаута (например, 90 секунд) Go аккуратно закроет соединение.
		TLSHandshakeTimeout:   2 * time.Second,  // Максимальное время, отведенное на TLS/SSL-рукопожатие. Установка защищенного HTTPS-соединения требует нескольких раундов обмена данными между серверами. Если удаленный сервер перегружен или сеть сильно тормозит, процесс генерации ключей может зависнуть. Этот таймаут прерывает запрос, если шифрование не настроилось за отведенное время (например, за 2 секунды).
		ExpectContinueTimeout: 1 * time.Second,  // Время, которое клиент ждет ответа 100 Continue от сервера после отправки HTTP-заголовков, но до начала отправки большого тела запроса (Request Body). Используется, только если вы отправляете заголовок Expect: 100-continue.
	},
}

func main() {
	urls := []string{
		"https://google.com1",
		"https://ya.ru",
	}

	fmt.Println(urls)

	// Контекст завершения системы
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Создаем дочерний контекст с таймаутом.
	// Ни одна горутина не должна выполняться дольше 3 секунд на весь цикл запроса.
	ctx, cancel := context.WithTimeout(rootCtx, 6*time.Second)
	// Timeout в http.Client будет установлен на 1 секунду меньше
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(urls))

	// Канал дял сбора ошибок
	errChan := make(chan error, len(urls))

	for _, v := range urls {
		go func(url string) {
			defer wg.Done()
			fmt.Println(url)
			// Простой способ
			// resp, err := http.Get(url)
			// // fmt.Println(resp)
			// // fmt.Println(err)

			// if err != nil {
			// 	fmt.Println("Error ", err)
			// 	return
			// }
			// // // Закрыть соединение
			// defer resp.Body.Close()

			// Другой способ
			body, err := fetchURL(ctx, url)
			if err != nil {
				errChan <- fmt.Errorf("ошибка для %s: %w", url, err)
			}

			// Результат отобразим в виде размера ответа
			fmt.Printf("%s: скачано %d байт\n", url, len(body))

		}(v)
	}

	// wg.Wait()

	// Важно: Ожидаем завершения всех горутин в фоне, чтобы не заблокировать чтение ошибок
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Читаем все ошибки, которые успели накопиться в канале
	for err := range errChan {
		if err != nil {
			fmt.Printf("[main err] %v\n", err)
		}
	}

	// Проверка завершения контекста
	if ctx.Err() != nil {
		fmt.Printf("работа завершена по контексту %v\n", ctx.Err())
	} else {
		fmt.Println("[main] Все запросы успешно обработаны.")
	}
}

func fetchURL(ctx context.Context, url string) ([]byte, error) {
	// Создаем запрос с привязкой к контексту приложения
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Выполняем запрос через оптимизированный клиент
	resp, err := customHTTPClient.Do(req)
	if err != nil {
		// Проверяем, не был ли запрос отменен извне через контекст
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, fmt.Errorf("request canceled by user/system: %w", ctx.Err())
		}
		return nil, fmt.Errorf("network error occurred: %w", err)
	}

	// Защита от утечки ресурсов. Body обязательно должно быть закрыто и вычитано.
	// Используем замыкание, чтобы не маскировать ошибку закрытия, или просто стандартный defer.
	defer func() {
		// Если не вычитать тело до конца, TCP-соединение нельзя будет переиспользовать (Keep-Alive)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	// Проверка HTTP-статуса (http.Get НЕ возвращает ошибку на статус 500 или 404)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Чтение тела с ограничением по размеру (Защита от Zip-бомб и гигантских ответов, забивающих RAM)
	maxBodySize := int64(1024 * 1024) // 1 MB
	limiterReader := io.LimitReader(resp.Body, maxBodySize)

	body, err := io.ReadAll(limiterReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// Глобальный http.Client: Создается один раз и переиспользуется. Если на каждый запрос делать new(http.Client) или вызывать http.Get(), приложение быстро исчерпает лимит доступных сокетов ОС (они будут висеть в состоянии TIME_WAIT)

// MaxIdleConnsPerHost: В http.DefaultClient этот параметр равен 2. Если ваш микросервис делает 100 запросов в секунду к одному и тому же API, 98 соединений будут закрываться и открываться заново, тратя ресурсы на TCP и TLS хэндшейки. Значение 100 решает эту проблему.

// io.Copy(io.Discard, resp.Body): Если вы прочитали только часть ответа (или вообще его не читали), перед закрытием тела (Body.Close()) Go-рантайм будет вынужден разорвать TCP-соединение. Если вычитать остатки в io.Discard, соединение останется живым и вернется в пул.

// io.LimitReader: Если сторонний сервер вместо ожидаемого JSON-файла на 1 КБ внезапно вернет файл размером 2 ГБ, стандартный io.ReadAll(resp.Body) загрузит его целиком в оперативную память. Сервис упадет по Out Of Memory (OOM). Ограничитель LimitReader предотвращает это.

// Обертывание ошибок (%w): Позволяет вызывающему коду на верхнем уровне делать errors.Is(err, context.Canceled) и понимать точную причину сбоя.
