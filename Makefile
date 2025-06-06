default: help

help:
	@echo "RIVA Community Outreach Bot Instructions"
	@echo
	@echo "    - make help             Print help information"
	@echo "    - make build            Build static binary"
	@echo "    - make build-image      Build container image"
	@echo "    - make run-image        Run container image"
	@echo "    - make publish-image    Push container image to docker.io"
	@echo "    - make clean            Clean any built binaries"
	@echo

.PHONY: build
build:
	CGO_ENABLED=1 CGO_LDFLAGS="-static" GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" go build -a -o rivabot -ldflags '-extldflags "-static" -w -s' .

.PHONY: build-image
build-image:
	podman build -t docker.io/taronaeo/rivabot -f Dockerfile --format docker .

.PHONY: run-image
run-image:
	podman run -d --name rivabot --restart always -v rivabot:/data docker.io/taronaeo/rivabot

.PHONY: publish-image
publish-image:
	podman push docker.io/taronaeo/rivabot

.PHONY: clean
clean:
	rm -f rivabot

