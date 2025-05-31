all: rivabot

rivabot:
	CGO_ENABLED=1 CGO_LDFLAGS="-static" GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" CXX="zig c++ -target x86_64-linux" go build -a -o rivabot -ldflags '-extldflags "-static" -w -s' .

clean:
	rm -f rivabot

