# Сбор и отправка метрик
Сбор рантайм-метрик (`cmd/agent`) с отправкой на сервер (`cmd/server`) по HTTP. Есть поддержка GZIP и простейшее шифрование.

## Сервер
Сервер может хранить данные в памяти и в Постгресе.

На сервере можно запустить механизм бэкапа из хранилища, он реализует задание по сохранению в файл. Бэкап не зависит от типа хранилища и может работать как с inmemory-базой так и с Постгресом.

Пример запуска (параметры описаны в `cmd/server/config/config.go`):

```sh
LOG_LEVEL=debug \
STORE_FILE=db.json \
KEY=secret \
STORE_INTERVAL=0s \
DATABASE_DSN=postgresql://localhost/praktikum_metrics \
go run cmd/server/server.go
```

## Агент
Агент хранит метрики в inmemory-базе и периодически отсылает их на сервер.

Пример запуска (параметры см. `cmd/agent/config/config.go`):

```sh
POLL_INTERVAL=1s \
LOG_LEVEL=debug \
KEY=secret \
REPORT_INTERVAL=2s \
go run cmd/agent/agent.go
```

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` - адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

# Обновление шаблона

Чтобы получать обновления автотестов и других частей шаблона, выполните следующую команду:

```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-devops-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.
