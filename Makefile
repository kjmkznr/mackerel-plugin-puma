TEST?=$$(go list ./... | grep -v '/vendor/')
VETARGS?=-all
TARGET=mackerel-plugin-puma
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: test vet

${TARGET}: prepare 
	go build

build: ${TARGET}

prepare:
	go get

clean:
	rm -Rf $(CURDIR)/bin/*

test: prepare vet
	go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4

vet: prepare 
	@echo "go tool vet $(VETARGS) ."
	@go tool vet $(VETARGS) $$(ls -d */ | grep -v vendor) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: default test vet
