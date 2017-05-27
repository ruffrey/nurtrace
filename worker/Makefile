preinstall:
	go get github.com/pkg/sftp
preinstall-test:
	go get github.com/ruffrey/nurtrace/potential
build:
	rm -f worker_linux*
	rm -f worker_darwin*
	GOOS=linux GOARCH=arm go build -ldflags "-s" -o worker_linux_arm
	GOOS=linux GOARCH=arm64 go build -ldflags "-s" -o worker_linux_arm64
	GOOS=linux GOARCH=amd64 go build -ldflags "-s" -o worker_linux_x86_64
	GOOS=linux GOARCH=386 go build -ldflags "-s" -o worker_linux_386
	GOOS=darwin GOARCH=386 go build -ldflags "-s" -o worker_darwin_386
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s" -o worker_darwin_x86_64
	chmod +x worker_*
