.PHONY: test test-python test-go test-ts lint up down build clean

test: test-python test-go test-ts

test-python:
	cd services/ingest-api && pip install -q -r requirements.txt && pytest -v

test-go:
	cd services/processor && go test -v ./...

test-ts:
	cd services/dashboard-bff && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd services/ingest-api && flake8 --max-line-length=120 --exclude=__pycache__ .

lint-go:
	cd services/processor && go vet ./...

lint-ts:
	cd services/dashboard-bff && npx eslint src/

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

clean:
	docker compose down -v --rmi local
