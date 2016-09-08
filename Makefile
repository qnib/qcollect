all: local alpine linux

test:
	gom test -cover ./...

local:
	./build.sh

alpine:
	docker run --rm -ti -v $(CURDIR):/data/ --workdir /data/ qnib/alpn-go-dev ./build.sh

linux:
	docker run --rm -ti -v $(CURDIR):/data/ --workdir /data/ qnib/golang ./build.sh
