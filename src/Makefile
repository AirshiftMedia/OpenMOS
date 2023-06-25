.PHONY: dev

dev:
go get gopkg.in/yaml.v2
go run main.go

.PHONY: run

run:
	@go run ./...

.PHONY: build

build:
	@go build -o ./build/server ./...
	@echo "[OK] Server build completed."

.PHONY: test

test:
	@go test -v -coverprofile=cover.out ./...
	@echo "[OK] Test and coverage file created."

.PHONY: show_coverage

show_coverage:
	@go tool cover -html=cover.out
	@echo "[OK] Coverage file opened."