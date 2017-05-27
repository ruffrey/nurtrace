preinstall:
	go get github.com/golang/glog
	go get github.com/stretchr/testify/assert
	go get github.com/y0ssar1an/q
	go get github.com/pkg/sftp
preinstall-test:
	go get github.com/ruffrey/nurtrace/laws

test:
	go test -v
test-race:
	go test -race bleh/potential

cover:
	go test -covermode=count -coverprofile=coverage.out # count how many times statements run
	go tool cover -html=coverage.out

profile:
	go test -v -cpuprofile=cprof
	go tool pprof -pdf potential.test cprof >> cprof.pdf
	open cprof.pdf
