# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


### Результат внесённых изменений

```
$ go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof

base.pprof profiles/result.pprof
File: server
Type: inuse_space
Time: 2026-01-30 01:23:40 +04
Showing nodes accounting for 518.63kB, 26.90% of 1927.64kB total
      flat  flat%   sum%        cum   cum%
  908.20kB 47.11% 47.11%   908.20kB 47.11%  github.com/goccy/go-json/internal/decoder.initDecoder.func1
 -902.59kB 46.82%  0.29%  -902.59kB 46.82%  compress/flate.NewWriter (inline)
     513kB 26.61% 26.90%      513kB 26.61%  runtime.allocm
  512.07kB 26.56% 53.47%   512.07kB 26.56%  github.com/go-chi/chi/v5.endpoints.Value (inline)
 -512.05kB 26.56% 26.90%  -512.05kB 26.56%  time.NewTicker
         0     0% 26.90%   512.07kB 26.56%  github.com/go-chi/chi/v5.(*Mux).Post (inline)
         0     0% 26.90%   908.20kB 47.11%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 26.90%   512.07kB 26.56%  github.com/go-chi/chi/v5.(*Mux).handle
         0     0% 26.90%   908.20kB 47.11%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 26.90%   512.07kB 26.56%  github.com/go-chi/chi/v5.(*node).InsertRoute
         0     0% 26.90%   512.07kB 26.56%  github.com/go-chi/chi/v5.(*node).setEndpoint
         0     0% 26.90%   908.20kB 47.11%  github.com/go-chi/chi/v5/middleware.(*Compressor).Handler-fm.(*Compressor).Handler.func1
         0     0% 26.90%  -902.59kB 46.82%  github.com/go-chi/chi/v5/middleware.(*Compressor).SetEncoder
         0     0% 26.90%  -902.59kB 46.82%  github.com/go-chi/chi/v5/middleware.Compress (inline)
         0     0% 26.90%  -902.59kB 46.82%  github.com/go-chi/chi/v5/middleware.NewCompressor
         0     0% 26.90%   908.20kB 47.11%  github.com/go-chi/chi/v5/middleware.StripSlashes.func1
         0     0% 26.90%  -902.59kB 46.82%  github.com/go-chi/chi/v5/middleware.encoderDeflate
         0     0% 26.90%   908.20kB 47.11%  github.com/go-chi/chi/v5/middleware.init.0.RequestLogger.func1.1
         0     0% 26.90%   908.20kB 47.11%  github.com/goccy/go-json.(*Decoder).Decode (inline)
         0     0% 26.90%   908.20kB 47.11%  github.com/goccy/go-json.(*Decoder).DecodeWithOption
         0     0% 26.90%   908.20kB 47.11%  github.com/goccy/go-json/internal/decoder.CompileToGetDecoder
         0     0% 26.90%   908.20kB 47.11%  github.com/goccy/go-json/internal/decoder.initDecoder
         0     0% 26.90%   908.20kB 47.11%  github.com/idudko/go-musthave-metrics/internal/handler.(*Handler).UpdateMetricsBatchHandler
         0     0% 26.90%   908.20kB 47.11%  github.com/idudko/go-musthave-metrics/internal/middleware.GzipRequestMiddleware.func1
         0     0% 26.90%   908.20kB 47.11%  github.com/idudko/go-musthave-metrics/internal/middleware.LoggingMiddleware.func1
         0     0% 26.90%  -512.05kB 26.56%  github.com/idudko/go-musthave-metrics/internal/repository.(*FileStorage).startAutoSave.func1
         0     0% 26.90%  -390.52kB 20.26%  main.main
         0     0% 26.90%  -390.52kB 20.26%  main.newServer
         0     0% 26.90%   908.20kB 47.11%  main.newServer.HashValidationMiddleware.func6.1
         0     0% 26.90%   908.20kB 47.11%  net/http.(*conn).serve
         0     0% 26.90%   908.20kB 47.11%  net/http.HandlerFunc.ServeHTTP
         0     0% 26.90%   908.20kB 47.11%  net/http.serverHandler.ServeHTTP
         0     0% 26.90%  -390.52kB 20.26%  runtime.main
         0     0% 26.90%      513kB 26.61%  runtime.mstart
         0     0% 26.90%      513kB 26.61%  runtime.mstart0
         0     0% 26.90%      513kB 26.61%  runtime.mstart1
         0     0% 26.90%      513kB 26.61%  runtime.newm
         0     0% 26.90%      513kB 26.61%  runtime.resetspinning
         0     0% 26.90%      513kB 26.61%  runtime.schedule
         0     0% 26.90%      513kB 26.61%  runtime.startm
         0     0% 26.90%      513kB 26.61%  runtime.wakep
         0     0% 26.90%   908.20kB 47.11%  sync.(*Once).Do (inline)
         0     0% 26.90%   908.20kB 47.11%  sync.(*Once).doSlow
```

### Ключевые оптимизации

1. **Кэширование ссылок на карты (`MemStorage`)**
   - Реализовано через `atomic.Value` для хранения ссылок
   - Убрано клонирование карт при чтении (`maps.Clone`)
   - **Результат**: 0 аллокаций при чтении независимо от размера карты

2. **Кэширование шаблонов (`Handler`)**
   - Шаблон HTML парсится один раз при создании Handler
   - Убрано динамическое создание строк шаблонов
   - **Результат**: Снижение аллокаций при рендеринге

3. **Оптимизация мьютексов**
   - Улучшена синхронизация с использованием `atomic.Int64` для отслеживания версий
   - Снижена блокировка при операциях чтения
   - **Результат**: Улучшение производительности при конкурентных чтениях
