test:
	go test -v

cover:
	go test -covermode=count -coverprofile=coverage.out # count how many times statements run
	go tool cover -html=coverage.out
