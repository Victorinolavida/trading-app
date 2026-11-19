module github.com/victorinolavida/gateway

go 1.26.5

require (
	github.com/go-chi/chi/v5 v5.5.0
	github.com/go-chi/cors v1.2.2
	github.com/go-playground/validator/v10 v10.30.5
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.12.1
	github.com/victorinolavida/auth v0.0.0
	github.com/victorinolavida/orders v0.0.0
	github.com/victorinolavida/pkg v0.0.0
	github.com/victorinolavida/proto v0.0.0
	github.com/victorinolavida/securities v0.0.0
	go.uber.org/goleak v1.3.0
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/flashlabs/rootpath v1.1.6 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/yaml v1.1.1 // indirect
	github.com/knadh/koanf/providers/env v1.1.0 // indirect
	github.com/knadh/koanf/providers/file v1.2.1 // indirect
	github.com/knadh/koanf/v2 v2.4.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/rabbitmq/amqp091-go v1.16.0 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	github.com/shopspring/decimal v1.6.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
)

replace github.com/victorinolavida/auth => ../users

replace github.com/victorinolavida/orders => ../orders

replace github.com/victorinolavida/pkg => ../pkg

replace github.com/victorinolavida/proto => ../gen

replace github.com/victorinolavida/securities => ../securities
