preinstall:
	go get gopkg.in/urfave/cli.v1
	go get github.com/awalterschulze/gographviz
	go get github.com/pkg/profile
preinstall-test:
	go get github.com/ruffrey/nurtrace/potential

build:
	go build -o nt

build-prod:
	go build -ldflags "-s -w" -o nt
