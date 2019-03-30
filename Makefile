test:
	go mod tidy -v
	go test -cover ./
