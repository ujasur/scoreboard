TARGET_OS?=linux
TARGET_ARCH?=amd64 
DEST?=./artifact

GOFILES = $(shell find . -maxdepth 1 -name '*.go')
DEPFILES = $(shell find . -maxdepth 1 -name 'go.*')
CONFIGFILES = $(shell find ./config -maxdepth 1 -name '*.*')
TEMPLATES = $(shell find ./templates -maxdepth 1 -name '*.*')
STATICFILES = $(shell find ./static -maxdepth 1 -name '*.*')
BINARY_NAME=scoreboard

default: workdir deps test compile clean
workdir:
	rm -rf $(DEST)	
	mkdir -p $(DEST)/config
	mkdir -p $(DEST)/templates
	mkdir -p $(DEST)/static
	cp $(GOFILES) $(DEST)
	cp $(DEPFILES) $(DEST)	
	cp $(CONFIGFILES) $(DEST)/config
	cp $(TEMPLATES) $(DEST)/templates	
	cp $(STATICFILES) $(DEST)/static
deps:
	cd $(DEST); go get -v -t -d ./...; 
test:
	cd $(DEST); go test -v ./...;
compile:
	cd $(DEST); GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -ldflags "-X main.version=$(VERSION)"  -o $(BINARY_NAME);
clean:
	rm $(DEST)/*.go