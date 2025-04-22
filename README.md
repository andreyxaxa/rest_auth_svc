# Часть сервиса аутентификации

## Используемые технологии:
- Go
- JWT
- PostgreSQL
- Docker

Payload токенов содержит сведения об IP, ID, email юзера.

Refresh token хранится в БД в виде bcrypt хеша.

## Маршруты:
- `POST /users` - создать юзера.

reqeust:
```json
{
    "email": "andrey123@gmail.com",
    "password": "password"
}
```
response:
```json
{
    "id": "1891ac8f-4d5e-4bb7-be4e-d903ca120213",
    "email": "andrey123@gmail.com"
}
```

- `POST /login` - залогинить юзера (поиск в БД + выдача пары JWT токенов).

request:
```json
{
    "email": "andrey123@gmail.com",
    "password": "password"
}
```
response:
```json
{
    "session_id": "f765bd2d-1b2b-42d1-a3ac-a27a7db8647b",
    "access_token":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE4OTFhYzhmLTRkNWUtNGJiNy1iZTRlLWQ5MDNjYTEyMDIxMyIsImVtYWlsIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJpcCI6Ils6OjFdOjUzNDU2Iiwic3ViIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJleHAiOjE3NDUzMzM4NDUsImlhdCI6MTc0NTMzMjk0NSwianRpIjoiZTk4ODMxYmItYjI3Ny00YzllLWJjYjMtMGYyZGY1MTQwNWFkIn0.q2blD1efRad84dS6HWzOp8Rx4etBuvIyIwoQcjy7wmTqNUyMA7knUuY11Ssz3BsiTv0TbkeLLEdEGy_WiDDewQ",
    "refresh_token":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE4OTFhYzhmLTRkNWUtNGJiNy1iZTRlLWQ5MDNjYTEyMDIxMyIsImVtYWlsIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJpcCI6Ils6OjFdOjUzNDU2Iiwic3ViIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJleHAiOjE3NDU0MTkzNDUsImlhdCI6MTc0NTMzMjk0NSwianRpIjoiZjc2NWJkMmQtMWIyYi00MmQxLWEzYWMtYTI3YTdkYjg2NDdiIn0.e5_fTIf5V2bwUh3pA30L6BKqYYw-SADkBRyy5vMcGR5DAUtWxPYthEqlVZhk3c93yHMo24tOMjSWvW0I_VyKbg",
    "access_token_expires_at":"2025-04-22T17:57:25+03:00",
    "refresh_token_expires_at":"2025-04-23T17:42:25+03:00",
    "user":
        {
          "id": "1891ac8f-4d5e-4bb7-be4e-d903ca120213",
          "email":"andrey123@gmail.com"
        }
}
```

- `GET /tokens/{id}` - выдача пары токенто по GUID юзера. Ответ такой же, как и в /login

- `POST /tokens/renew` - обновление access token'а. Если IP-адрес клиента изменился - посылается email-warning на почту юзера.

request:
```json
{
    "refresh_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjU4YTNjNDAyLThmYzgtNDk3MS05ODI2LWQxMDQ4NTc3NzRiMiIsImVtYWlsIjoibjFpa2l0YTEyM0BnbWFpbC5jb20iLCJpcCI6Ils6OjFdOjY1NDAxIiwic3ViIjoibjFpa2l0YTEyM0BnbWFpbC5jb20iLCJleHAiOjE3NDUzNjYxMzEsImlhdCI6MTc0NTI3OTczMSwianRpIjoiNTA0ZGNmMDYtODNjZS00YWIxLWE0Y2UtYjFlNmU2NWEyNDUxIn0.CsTnVzXWigfbJthVPaMO6yK90IIqjrowMrDmBr5a5dgXV5vuY2_TYi-rfx1yXsnH2-kR8YnMANQdE8n6APDRKA"
}
```
response:
```json
{
    "access_token":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjE4OTFhYzhmLTRkNWUtNGJiNy1iZTRlLWQ5MDNjYTEyMDIxMyIsImVtYWlsIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJpcCI6Ils6OjFdOjUzNDU2Iiwic3ViIjoiYW5kcmV5MTEyM0BnbWFpbC5jb20iLCJleHAiOjE3NDUzMzM4NDUsImlhdCI6MTc0NTMzMjk0NSwianRpIjoiZTk4ODMxYmItYjI3Ny00YzllLWJjYjMtMGYyZGY1MTQwNWFkIn0.q2blD1efRad84dS6HWzOp8Rx4etBuvIyIwoQcjy7wmTqNUyMA7knUuY11Ssz3BsiTv0TbkeLLEdEGy_WiDDewQ",
    "access_token_expires_at":"2025-04-22T17:57:25+03:00"
}
```

- `POST /tokens/refresh` - рефреш пары токенов. В запрос id юзера и id сессии

request:
```json
{
    "id": "1891ac8f-4d5e-4bb7-be4e-d903ca120213",
    "session_id": "f765bd2d-1b2b-42d1-a3ac-a27a7db8647b"
}
```
response: такой же, как и в /login
