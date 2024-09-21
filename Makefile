APP = mediadownload

build:
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o $(APP)

clean:
	@rm -rf $(APP)

.PHONY: build clean