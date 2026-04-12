# Подробная инструкция для мобильщика: wwvpn

## 1) Актуальные параметры стенда

- API base URL: `http://64.176.73.163:8080/api/v1`
- Health-check: `http://64.176.73.163:8080/health`
- Тестовый `device_id`: `mobile-test-device-1774298753`
- Формат авторизации: `Authorization: Bearer <token>`
- Рабочий `server_id`: `8460b06b-a811-4266-9e52-e0b7cda92c51`

Если этот `device_id` занят, можно использовать один из batch:
- `mobile-batch-1774300382-01`
- `mobile-batch-1774300382-02`
- `mobile-batch-1774300382-03`
- `mobile-batch-1774300382-04`
- `mobile-batch-1774300382-05`
- `mobile-batch-1774300382-06`
- `mobile-batch-1774300382-07`
- `mobile-batch-1774300382-08`

## 2) End-to-end flow

1. `POST /auth/device` -> получить JWT
2. `GET /subscription/status` -> убедиться, что `status=active`
3. `GET /servers` -> выбрать активный сервер
4. Сгенерировать клиентскую пару ключей (private/public)
5. `POST /connect` с `client_public_key`
6. Применить VPN-конфиг в системный клиент
7. Проверить интернет через VPN
8. После теста вызвать `POST /connect/disconnect`

## 3) API примеры

### 3.1 Авторизация устройства

`POST /api/v1/auth/device`

```json
{
  "device_id": "mobile-test-device-1774298753"
}
```

Ответ содержит:
- `token`
- `expire_at`
- `device_id`

### 3.2 Статус подписки

`GET /api/v1/subscription/status`

Headers:
- `Authorization: Bearer <token>`

Ожидание:
- `status: active`

### 3.3 Список серверов

`GET /api/v1/servers`

Headers:
- `Authorization: Bearer <token>`

Ожидание:
- `servers` не пустой
- есть `id`, `country`, `city`, `protocols`

### 3.4 Connect

`POST /api/v1/connect`

Headers:
- `Authorization: Bearer <token>`
- `Content-Type: application/json`

```json
{
  "server_id": "8460b06b-a811-4266-9e52-e0b7cda92c51",
  "protocol": "amnezia_wg",
  "client_public_key": "PUT_CLIENT_PUBLIC_KEY_HERE"
}
```

`protocol`:
- `amnezia_wg`
- `vless_reality`
- `auto`

Для стабильного теста сейчас используй `amnezia_wg`.

### 3.5 Disconnect

`POST /api/v1/connect/disconnect`

Headers:
- `Authorization: Bearer <token>`
- `Content-Type: application/json`

```json
{
  "connection_id": "PUT_CONNECTION_ID_FROM_CONNECT_HERE"
}
```

## 4) Что обязательно применить на клиенте

После `connect` сервер возвращает:
- `endpoint`
- `server_public_key`
- `assigned_ip`
- `allowed_ips`
- `dns`
- `mtu`
- `obfuscation` (`jc`, `jmin`, `jmax`, `s1`, `s2`, `h1`, `h2`, `h3`, `h4`)

Критично:
- в интерфейс клиента должен попасть именно `privateKey`, соответствующий `client_public_key`, который ушел в `/connect`
- `server_public_key` должен быть строго из ответа API
- для AmneziaWG должны применяться параметры `obfuscation`

## 5) Проверка, что интернет реально идет

1. До подключения открыть `https://ifconfig.me`
2. Поднять туннель
3. Снова открыть `https://ifconfig.me`
4. Если IP сменился на IP VPS, туннель и интернет работают

Дополнительно:
- `https://dnsleaktest.com` для проверки DNS

## 6) Диагностика частых ошибок

### `subscription limit reached`

Причина:
- для device превышен лимит активных соединений

Решение:
- сделать `/connect/disconnect` для старого `connection_id`
- либо взять новый `device_id`

### `wg set failed: Key is not the correct length or format`

Причина:
- в `client_public_key` передана строка, не WG/AWG public key

Решение:
- отправлять только валидный base64 public key от X25519

### Подключено, но интернета нет

Причина:
- на сервер UDP доходит, но handshake AWG не проходит

Проверить на клиенте:
- что `client_public_key` в `/connect` соответствует `privateKey`, который реально применен
- что применены все поля `obfuscation`
- что используется именно AmneziaWG-движок, а не обычный WG без obfuscation

## 7) Мини-чеклист перед запуском теста

1. Новый или свободный `device_id`
2. `subscription/status = active`
3. `servers` не пустой
4. Валидная пара ключей и консистентность private/public
5. `connect` возвращает конфиг
6. Туннель поднялся, IP сменился
7. После теста сделан `disconnect`

