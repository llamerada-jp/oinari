SHELL := /bin/bash -o pipefail

COLONIO_BRANCH := main
COLONIO_FILES := dist/colonio.js dist/colonio_go.js dist/colonio.wasm
GO_FILES := $(shell find . -name *.go | grep -v ./build/)
TS_FILES := $(shell find ./src/ -name *.ts) src/colonio.d.ts src/colonio_go.d.ts src/keys.ts
OINARI_FILES := dist/wasm_exec.js dist/404.html dist/error.html
MAP_FILES := dist/colonio.wasm.map dist/colonio_go.js.map

.PHONY: build
build: $(COLONIO_FILES) $(OINARI_FILES) build-go build-ts

.PHONY: build-go
build-go: $(GO_FILES) go.mod go.sum
	git show --format='%H' --no-patch > ./cmd/seed/commit_hash.txt
	go build -o ./bin/seed ./cmd/seed
	git show --format='%H' --no-patch > ./cmd/node/commit_hash.txt
	GOOS=js GOARCH=wasm go build -o ./dist/oinari.wasm ./cmd/node
	GOOS=js GOARCH=wasm go build -o ./dist/test/exit.wasm ./cmd/app/exit
	GOOS=js GOARCH=wasm go build -o ./dist/test/fox.wasm ./cmd/app/fox
	GOOS=js GOARCH=wasm go build -o ./dist/test/sleep.wasm ./cmd/app/sleep
	GOOS=js GOARCH=wasm go build -o ./dist/test/sleep_core.wasm ./cmd/app/sleep_core
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_crosslink.wasm -c ./lib/crosslink/
	## should edit TESTS@src/test.ts to run test build by wasm
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_api_core.wasm -c ./api/core/
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_node.wasm -c ./cmd/node/

.PHONY: build-ts
build-ts: $(TS_FILES) package.json tsconfig.json webpack.config.js
	npm run build
	mv ./dist/test.js ./dist/test/test.js

.PHONY: setup
setup:
	mkdir -p build
	# colonio
	rm -rf build/colonio
	if [ "$${COLONIO_DEV_PATH}" = "" ]; \
	then git clone -b $(COLONIO_BRANCH) --depth=1 https://github.com/llamerada-jp/colonio.git build/colonio; \
	else ln -s $${COLONIO_DEV_PATH} build/colonio; \
	fi
	$(MAKE) -C build/colonio build BUILD_TYPE=Debug
	npm install	

.PHONY: s
s: build generate-cert
	sudo sysctl -w net.core.rmem_max=2500000
	while true; do go run ./cmd/seed --debug; done

.PHONY: test
test: build generate-cert
	sudo sysctl -w net.core.rmem_max=2500000
	npm t
	go run ./cmd/seed --test

dist/404.html: src/404.html secrets.json
	go run ./cmd/tool template -i src/404.html -v secrets.json > $@
  
dist/colonio.js: build/colonio/output/colonio.js
	cp $< $@

dist/colonio.wasm: build/colonio/output/colonio.wasm
	cp $< $@

dist/colonio.wasm.map: build/colonio/output/colonio.wasm.map
	cp $< $@

dist/colonio_go.js: build/colonio/src/js/colonio_go.js
	cp $< $@

dist/colonio_go.js.map: build/colonio/src/js/colonio_go.js.map
	cp $< $@

dist/error.html: src/error.html secrets.json
	go run ./cmd/tool template -i src/error.html -v secrets.json > $@

dist/wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< $@

src/colonio.d.ts: build/colonio/src/js/core.d.ts
	cp $< $@

src/colonio_go.d.ts: build/colonio/src/js/colonio_go.d.ts
	cp $< $@

src/keys.ts: src/keys.temp secrets.json
	go run ./cmd/tool template -i src/keys.temp -v secrets.json > $@

.PHONY: generate-cert
generate-cert:
	openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

.PHONY: clean
clean:
	rm -f ./dist/*.js ./dist/**/*.js ./dist/*.map ./dist/**/*.map ./dist/*.wasm ./dist/**/*.wasm $(OINARI_FILES) ./src/colonio_go.d.ts ./src/colonio.d.ts ./src/keys.ts localhost.crt localhost.key

.PHONY: deisclean
deisclean: clean
	rm -fr ./node_modules ./coverage ./build