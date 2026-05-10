module github.com/tvvshow/gokao/pkg/database

go 1.23.0

require (
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.9
	github.com/tvvshow/gokao/pkg/config v0.0.0
	github.com/redis/go-redis/v9 v9.13.0
	github.com/sirupsen/logrus v1.9.3
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.30.2
)

replace github.com/tvvshow/gokao/pkg/config => ../config

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
