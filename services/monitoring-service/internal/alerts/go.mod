module github.com/tvvshow/gokao/services/monitoring-service/internal/alerts

go 1.23.0

require (
	github.com/redis/go-redis/v9 v9.13.0
	github.com/tvvshow/gokao/services/monitoring-service/internal/metrics v0.0.0-20260511033515-55fe82e1d0e7
	go.uber.org/zap v1.27.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
)

replace github.com/tvvshow/gokao/services/monitoring-service/internal/metrics => ../metrics
