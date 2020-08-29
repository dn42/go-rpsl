-include secret.inc

ID=registry-go-rpsl-coverage

test:
	go vet ./...
	go test -race -cover -coverprofile=c.out -covermode atomic ./...
	go tool cover -func=c.out | tee func.txt

report:
	./coverage.sh func.txt
