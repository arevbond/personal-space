.PHONY: test-cover

run:
	docker compose -f compose.dev.yml up
restart:
	docker compose -f compose.dev.yml restart app
test-cover:
	rm -rf test-cover
	mkdir test-cover
	go test -coverpkg=./internal/... -coverprofile=test-cover/coverage.out ./...
	go tool cover -html=test-cover/coverage.out -o test-cover/coverage.html