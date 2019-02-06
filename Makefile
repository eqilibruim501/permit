.PHONY: build realize deps fmt vet qa clean

include .env

BASE_DIR = github.com/crusttech/permit
PKG = ${BASE_DIR}

GOBUILD   = CGO_ENABLED=0 go build
GORUN     = go run
GOGET     = go get
GOGEN     = go generate
GODEP     = ${GOPATH}/bin/dep
REALIZE   = ${GOPATH}/bin/realize

build:
	mkdir -p build
	$(GOBUILD) ${LDFLAGS} -o build/api ${PKG}/cmd/cli

realize: $(REALIZE)
	$(REALIZE) start

deps: $(GODEP) $(GOMIGRATE)
	$(GODEP) ensure -v

$(GODEP):
	$(GOGET) github.com/golang/dep/cmd/dep

$(REALIZE):
	$(GOGET) github.com/tockins/realize

fmt:
	@echo `${PATHS_GOSRC}`| xargs -n1 gofmt -e -l

vet:
	go vet `cd ${GOPATH}/src/; find $(PKG)/internal $(PKG)/cmd -type f -name '*.go' -and -not -path '*vendor*'|xargs -n1 dirname|uniq`

qa: fmt vet test

clean:
	rm -rf build
	go clean
