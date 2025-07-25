module github.com/KromaEnergia/api-consultor

go 1.23.0

toolchain go1.23.10

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/gorilla/mux v1.8.1
	golang.org/x/crypto v0.39.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.0

)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/joho/godotenv v1.5.1
	github.com/rs/cors v1.11.1
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace github.com/KromaEnergia/api-consultor => ./
