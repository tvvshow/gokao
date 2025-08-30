.PHONY: help go-setup go-run go-test fmt

help:
@echo \"Targets:\"
@echo \"  go-setup   - go mod tidy in services/api-gateway\"
@echo \"  go-run     - run api-gateway on :8080\"
@echo \"  go-test    - run go tests (if any)\"
@echo \"  fmt        - format go code\"

go-setup:
cd services/api-gateway && go mod tidy

go-run:
cd services/api-gateway && go run .

go-test:
cd services/api-gateway && go test ./... -cover

fmt:
gofmt -w services/api-gateway