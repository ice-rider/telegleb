# Telegleb

Веб-клиент Telegram: Go-бэкенд поверх MTProto (`gotd/td`) + SolidJS-фронтенд. Пользователь логинится
своим номером телефона, бэкенд держит живой MTProto-клиент на каждую сессию и проксирует чаты,
сообщения и медиа в браузер.

## Документация

- [`docs/API.md`](docs/API.md) — **нормативная спецификация контракта v1**. Единственный источник
  правды по формату обмена. Любое изменение формата вносится сначала туда, потом в обе реализации;
  менять DTO только на одной стороне запрещено.
- [`docs/AUDIT.md`](docs/AUDIT.md) — разбор дефектов, из-за которых контракт переписывали.
  Объясняет, почему спека устроена именно так, и перечисляет незакрытое.

## Структура

```
server/   Go 1.26, fasthttp + gotd/td + Redis, DI через google/wire
client/   SolidJS + Vite + TypeScript + axios
proxy/    nginx: /api/ → backend:8080, / → frontend:80
```

## Команды

```bash
# backend
cd server && go build ./...
cd server && go test ./...
cd server && go run ./cmd/app          # нужен Redis + TG_APP_ID/TG_APP_HASH/JWT_SECRET

# frontend
cd client && npm run dev               # vite, проксирует /api на localhost:8080
cd client && npm run build             # tsc -b && vite build
cd client && npx tsc -b                # только тайпчек

# всё вместе
make up / make down / make logs        # docker compose (dev, со сборкой)
make prod-up / make prod-down          # образы из ghcr.io/ice-rider/telegleb
make build / make push-all             # сборка и пуш образов в GHCR
```

Wire генерируется: `cd server && go run -mod=mod github.com/google/wire/cmd/wire ./internal/app`.
Провайдеры должны отдавать различимые типы — четыре голых параметра одного типа wire развести не
может, поэтому параметры Telegram собраны в `telegram.Config`.

## Главный инвариант: чат не рисуется без своей папки

`chat.archived` и `chat.folderIds` вычисляет сервер, и они лежат **на самом чате**, а не в
отдельном справочнике папок. `GET /chats` отдаёт чаты, папки и профиль одним ответом; клиент
держит дашборд одним значением состояния и не рисует ни одного чата, пока ответ не получен
целиком.

Это защита от конкретного дефекта: схема «чаты отдельно, папки отдельно» допускает кадр, где чаты
уже есть, а раскладка ещё нет — официальный клиент Telegram именно так на мгновение показывает
архивные чаты в общем списке. Когда размещение приезжает тем же объектом, такого кадра не бывает.

Отсюда правила, которые нельзя нарушать:
- фильтрация на клиенте читает поля чата, а не соединяет два списка;
- если хоть одна часть ответа не загрузилась, весь ответ неуспешен — половина раскладки хуже
  индикатора загрузки;
- то же распространится на локальный кэш, если он появится.

Архив доступен клиенту отдельной вкладкой, но во вкладку «Все» его чаты не попадают никогда.

## Архитектура бэкенда (clean architecture)

```
cmd/app/main.go            → app.InitApp() (wire) → App.Run()
internal/core/domain/      модели + Peer/MediaRef с их разбором (peer.go, message.go)
internal/core/usecase/     юзкейсы + интерфейсы репозиториев (порты)
  auth/                    request-code, verify-code, verify-password, session, logout
  messenger/               dashboard, история, отправка, стриминг медиа
  chat|message|media|session/  только интерфейсы репозиториев
internal/adapter/telegram/ реализация портов через gotd/td
  chat.go                  GetDashboard: диалоги + архив + фильтры + профиль
  folders.go               раскладка по папкам (предикаты, включения, исключения)
  mapper.go                словари сущностей и маппинг сообщений
internal/adapter/repository/session/redis.go  AuthSession в Redis (ключ = сам токен)
internal/delivery/http/    fasthttp: router.go, handler_*.go, dto.go, errors.go
```

Направление зависимостей: delivery → usecase → domain; adapter реализует интерфейсы из usecase.

### Ключевые детали

- **`accessHash` обязателен.** Адрес чата — `domain.Peer` и его строковый `ref`
  (`user:<id>:<hash>` / `chat:<id>` / `channel:<id>:<hash>`). Собирать `InputPeer` из одного
  числового id нельзя — Telegram ответит `PEER_ID_INVALID`.
- **Карта сообщений ключуется парой (чат, id).** Нумерация уникальна только внутри чата, у
  каналов начинается с единицы; плоский ключ по `msg.ID` подставлял в диалог чужое превью.
- **Время сериализуется только через `t.UTC().Format(time.RFC3339)`.** Раскладка
  `"2006-01-02T15:04:05Z"` трактует `Z` как литерал и оставляет время локальным.
- **Пул MTProto-клиентов** — `map[sessionToken]*ActiveClient`, каждый клиент в своей горутине с
  долгоживущим `client.Run(...)`. Состояние сессии пишется в Redis через `SessionBridge`.
  Пул живёт в памяти процесса — бэкенд **не масштабируется горизонтально** без sticky-сессий.
- **`fileReference` не сохраняется.** `ref` медиа указывает на сообщение, сервер перезапрашивает
  его и берёт свежий reference в момент скачивания.

## Фронтенд

- SolidJS, сигналы объявлены на уровне модуля в `features/*/store.ts` — глобальные синглтоны,
  `useXxx()` возвращает к ним доступ. Контекста/провайдеров нет.
- Имена полей совпадают с JSON один в один, поэтому слоя переименования нет: `src/core/api/client.ts`
  только ходит в сеть и разворачивает ошибки в `ApiError` с машинным `code`.
- Разлогин по `SESSION_EXPIRED` живёт в перехватчике axios, а не в вызовах. При старте токен
  проверяется через `GET /auth/session`.
- Своё сообщение определяется флагом `message.out`.
- Открытие чата привязано к действию, а не к монтированию; `<Show keyed>` в `App.tsx`
  дополнительно пересоздаёт `ChatWindow` при смене чата.
- Алиас `~` → `client/src` (vite + tsconfig paths), `strict` включён.
- Стили: обычный CSS, по файлу рядом с компонентом, БЭМ-подобные имена. UI на русском.

## Конфигурация (env, cleanenv; читается `.env`, иначе окружение)

Обязательные: `TG_APP_ID`, `TG_APP_HASH`, `JWT_SECRET`. Шаблон — `server/.env.example`.
Прочие: `APP_ENV`, `HTTP_PORT` (8080), `LOG_LEVEL`, `LOG_JSON`, `JWT_TTL` (72h),
`REDIS_HOST/PORT/PASSWORD/DB/TIMEOUT`, `TG_PROXY_ADDR` + `TG_PROXY_SECRET` (MTProxy; иначе
прямое подключение через `dcs.Plain`).

## Состояние

Оба таргета собираются чисто, `go test ./...` зелёный. CORS убран — фронт и API одного
происхождения и в проде (nginx), и в деве (прокси vite).

Незакрытое: аватары, отметка прочитанного, live-обновления, пагинация в UI, sticky-сессии.
Подробности — в конце [`docs/AUDIT.md`](docs/AUDIT.md).

## Деплой

Push в `main`/`dev` → GitHub Actions собирает три образа (backend/frontend/proxy) и пушит в
`ghcr.io/ice-rider/telegleb/*` с тегами `latest` и коротким SHA. Прод разворачивается
`docker-compose.prod.yaml` (в нём плейсхолдеры `CHANGE_ME` для секретов).
