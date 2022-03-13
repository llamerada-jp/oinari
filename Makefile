
COLONIO_BRANCH := add-headers-option
COLONIO_FILES := src/colonio.js dist/colonio.wasm
GO_FILES := $(shell find . -name *.go)
OINARI_FILES := dist/wasm_exec.js

run: build
	while true; do ./bin/seed --debug -p 8080; done

.PHONY: setup
setup:
	mkdir -p build
	# colonio
	rm -rf build/colonio
	git clone -b $(COLONIO_BRANCH) --depth=1 https://github.com/llamerada-jp/colonio.git build/colonio
	$(MAKE) -C build/colonio build
	npm install

build: $(COLONIO_FILES) $(GO_FILES) $(OINARI_FILES) bin/seed dist/index.html src/keys.ts
	GOOS=js GOARCH=wasm go build -o dist/oinari.wasm ./cmd/agent/*.go
	npm run build

test:
	npm t

bin/seed: $(GO_FILES)
	go build -o $@ ./cmd/seed

dist/colonio.wasm: build/colonio/output/colonio.wasm
	cp $< $@

dist/index.html: src/index.html keys.json
	go run ./cmd/tool template -i src/index.html -v keys.json > $@

dist/wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< $@

src/colonio.js: build/colonio/output/colonio.js
	cp $< $@

src/keys.ts: src/keys.temp keys.json
	go run ./cmd/tool template -i src/keys.temp -v keys.json > $@
