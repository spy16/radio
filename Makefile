test:
	go mod tidy -v
	go test -cover ./.

test-verbose:
	go test -count=1 -v -cover ./.
