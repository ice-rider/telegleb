# Развёртывание

Прод крутится на одной машине в docker compose: `redis`, `backend`, `frontend`
и `proxy` (nginx, терминирует TLS). Схема: `proxy` → `/api/` в `backend:8080`,
всё остальное во `frontend:80`.

## Что лежит на сервере

```
/opt/telegleb/docker-compose.yaml   копия docker-compose.prod.yaml
/opt/telegleb/.env                  секреты, chmod 600, в git не попадает
/etc/telegleb/certs/fullchain.crt   сертификат + промежуточные CA
/etc/telegleb/certs/private.key     приватный ключ, chmod 600
```

Сертификат монтируется в контейнер прокси только на чтение. В образ он не
запекается: обновление сертификата не требует пересборки, а ключ не остаётся
в слоях, которые можно случайно куда-нибудь запушить.

## Секреты

Шаблон — `deploy/env.example`. `TG_APP_ID` и `TG_APP_HASH` берутся на
https://my.telegram.org → API development tools. `JWT_SECRET` генерируется:

```bash
openssl rand -base64 48
```

## Сборка и выкладка образов

На машине с 1 CPU и 1 ГБ памяти сборка фронта (vite) ненадёжна, поэтому образы
собираются локально и переносятся готовыми:

```bash
docker build -t ghcr.io/ice-rider/telegleb/backend:latest ./server
docker build -t ghcr.io/ice-rider/telegleb/frontend:latest ./client
docker build -t ghcr.io/ice-rider/telegleb/proxy:latest ./proxy

docker save ghcr.io/ice-rider/telegleb/{backend,frontend,proxy}:latest \
  | gzip -1 > images.tar.gz
scp images.tar.gz root@<host>:/opt/telegleb/
ssh root@<host> 'cd /opt/telegleb && docker load -i images.tar.gz && rm images.tar.gz'
```

Имена совпадают с тегами GHCR, а `pull_policy: never` в compose не даёт
подтянуть поверх устаревший `:latest` из реестра.

## Запуск

```bash
cd /opt/telegleb
docker compose up -d
docker compose ps
docker compose logs -f backend
```

Состояние проверяется через `GET /api/v1/health` — он же используется как
healthcheck контейнера.

## Обновление сертификата

```bash
cat certificate.crt certificate_ca.crt > fullchain.crt   # порядок важен
scp fullchain.crt certificate.key root@<host>:/etc/telegleb/certs/
ssh root@<host> 'chmod 600 /etc/telegleb/certs/private.key
                 docker compose -f /opt/telegleb/docker-compose.yaml restart proxy'
```

Перед заливкой стоит убедиться, что ключ соответствует сертификату:

```bash
diff <(openssl x509 -in fullchain.crt -noout -modulus) \
     <(openssl rsa  -in certificate.key -noout -modulus) && echo "пара сходится"
```

## О заголовках безопасности

`add_header` внутри `location` отменяет **все** заголовки, унаследованные из
`server`. Из-за этого HSTS и CSP раньше молча не отдавались. Поэтому в
`client/nginx.conf` заголовки заданы одним блоком на уровне `server`, а в
блоке статики, где нужен свой `Cache-Control`, нужные заголовки повторены явно.

HSTS выставляется на границе TLS (`proxy/nginx.conf`) и намеренно **без**
`includeSubDomains`: в SAN сертификата есть `mail`, `owa` и `autodiscover`,
и принудительный HTTPS сломал бы их, если они обслуживаются по http.
