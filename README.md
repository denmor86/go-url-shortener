# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

### Добавление информации о сборке
```
go run -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(Get-Date -Format "yyyy/MM/dd HH:mm:ss")' -X 'main.buildCommit=$(git rev-parse --short HEAD)'" main.go
```

### Генерация GRPC API
```
protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import -I internal/proto/ internal/proto/shortener.proto
```

### Запуск нагрузочного тестирования
```
.\cmd\benchmark\benchmark.exe
```

### Запуск тестов
```
cd script && python.exe .\run_tests.py --increments=25 | Select-String "FAIL" 
```

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
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
