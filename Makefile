run:
	docker compose -f compose.yml up
restart:
	docker compose -f compose.yml restart app
cover:
	go test -coverprofile=test-cover/coverage.out ./... &&	go tool cover -html=test-cover/coverage.out -o test-cover/coverage.html
