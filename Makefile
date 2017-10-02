
COMMANDS = dicot-api dicot-pwhash

BINARIES = $(COMMANDS:%=bin/%s)

all: binaries conf

binaries: .vendor.status
	go install ./pkg/...
	go install ./cmd/...
	mkdir -p bin/
	for c in $(COMMANDS); \
	do \
		rm -f ./bin/$$c; \
		ln -s $$GOPATH/bin/$$c ./bin/$$c; \
	done

$(BINARIES): binaries

conf/admin-password.txt:
	dd if=/dev/random bs=1 count=20 2>/dev/null | base64 > $@

conf/identity_admin: conf/identity_admin.in conf/admin-password.txt
	PW=`cat conf/admin-password.txt` && \
		sed -e "s,::ADMIN-PASSWORD::,$${PW}," < $< > $@ || rm $@

manifests/050-identity-project.yaml: manifests/050-identity-project.yaml.in conf/admin-password.txt  bin/dicot-pwhash
	PW=`bin/dicot-pwhash --password-file=conf/admin-password.txt` && \
		sed -e "s,::ADMIN-PASSWORD::,$${PW}," < $< > $@ || rm $@

conf: manifests/050-identity-project.yaml conf/identity_admin

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
