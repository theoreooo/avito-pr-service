# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы.

## Запуск

### С помощью Docker Compose

```bash
docker-compose up --build
```

Сервис будет доступен на `http://localhost:8080`
Swagger документация `http://localhost:8080/swagger/index.html`

### Makefile команды

```bash
make gen       # Сгенерировать API код из openapi.yaml
make up        # Запустить контейнеры (docker-compose up --build)
make down      # Остановить контейнеры
make clean     # Очистить сгенерированные файлы
```

## API Endpoints

### Teams
- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=...` - Получить команду

### Users
- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=...` - Получить PR'ы пользователя в роли ревьювера

### Pull Requests
- `POST /pullRequest/create` - Создать PR (автоназначаются ревьюверы)
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить ревьювера

### Statistics
- `GET /statistics` - Получить статистику по PR и ревьюверам


```

## Архитектура
handler/          # HTTP-слой (chi + oapi-codegen)
service/          # бизнес-логика (PRService, TeamService, UserService)
repository/       # работа с PostgreSQL
PostgreSQL

## Сделанные Допущения и Решения

### Могут ли существовать пустые команды?
Да, могут. Все участники команды могут покинуть её. При этом если попытаться получить команду через /get, то выйдет сообщение о том, что команды не существует, но сама запись о пустой команде останется в таблице и в нее могут вернуться пользователи.

### Вместо того чтобы выносить сложную логику выбора кандидатов и управления транзакциями на cервисный уровень, z инкапсулировал её в PRRepository.
Все сложные операции (CreatePR, ReassignReviewer, MergePR) выполняются в рамках одной PostgreSQL транзакции.

```