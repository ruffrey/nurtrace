test:
	go test -v

cover:
	go test -covermode=count -coverprofile=coverage.out # count how many times statements run
	go tool cover -html=coverage.out

profile:
	go test -v -cpuprofile=cprof
	go tool pprof -pdf potential.test cprof >> cprof.pdf
	open cprof.pdf
