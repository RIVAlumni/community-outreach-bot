all: build

build:
	CGO_ENABLED=1 CGO_LDFLAGS="-static" GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" go build -a -o rivabot -ldflags '-extldflags "-static" -w -s' .

build-image:
	podman build -t docker.io/taronaeo/rivabot -f Dockerfile --format docker .

run-image:
	podman run -d --name rivabot --restart always -v rivabot:/data docker.io/taronaeo/rivabot

publish-image:
	podman push docker.io/taronaeo/rivabot

clean:
	rm -f rivabot

