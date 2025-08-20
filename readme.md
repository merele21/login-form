# Project structure

```
login-form/
├── protos/ # proto файлы
    ├── gen/ # generator
        ├── go/
            ├── sso/
                ├── sso.pb.go
                └── sso_grpc.pb.go
    ├── proto
        ├── sso
            └── sso.proto
├── sso # название gRPC проекта
    ├── cmd/
        ├── migrator/
            └── main.go
        ├── sso/
            └── main.go
    ├── config/ # конфигурации
        └── local.yaml
    ├── internal/ # основной код
        ├── app/
            ├── grpc/
                └── app.go
            └── app.go
        ├── config/
            └── config.go
        ├── domain/
            └── models/
                ├── app.go
                └── user.go
        ├── grpc/
            └── auth/
                └── server.go
        ├── lib/
            ├── jwt/
                └── jwt.go
            └── logger/
        ├── services/
            ├── auth/
                └── auth.go
            └── permissions/
        ├── storage # БД
            ├── sqlite # (заменить на postgresql)
                └── sqlite.go # (заменить на postgres.go)
            └── storage.go
    ├── migrations/ # миграции
        ├── 1_init.down.sql
        ├── 1_init.up.sql
        ├── 2_add_is_admin_column_to_users_tbl.down.sql
        └── 2_add_is_admin_column_to_users_tbl.up.sql
    ├── storage/
    ├── tests # тесты (юнит/функциональные/интеграционные)
```