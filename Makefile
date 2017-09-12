
COMMANDS = dicot-api

BINARIES = $(COMMANDS:%=bin/%s)

all: binaries

binaries: .vendor.status
	go install ./pkg/...
	go install ./cmd/...
	mkdir -p bin/
	for c in $(COMMANDS); \
	do \
		rm -f ./bin/$$c; \
		ln -s $$GOPATH/bin/$$c ./bin/$$c; \
	done

.vendor.status: glide.yaml glide.lock
	glide install --strip-vendor && touch .vendor.status

glide-update:
	rm -rf glide.lock .vendor.status vendor
	glide cc
	glide update --strip-vendor && touch .vendor.status

clean:
	rm -rf vendor
	rm -rf bin/
	rm -rf $$GOPATH/pkg/*/github.com/dicot-project/dicot-api
	rm -f $$GOPATH/bin/dicot-api
