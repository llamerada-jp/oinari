
COLONIO_BRANCH := main
COLONIO_FILES := dist/colonio.js dist/colonio_go.js dist/colonio.wasm
GO_FILES := $(shell find . -name *.go)
OINARI_FILES := dist/wasm_exec.js dist/404.html dist/error.html dist/index.html
MAP_FILES := dist/colonio.wasm.map dist/colonio_go.js.map

run: build test-wasm
	while true; do ./bin/seed --debug -p 8080; done

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

.PHONY: build
build: $(COLONIO_FILES) $(GO_FILES) $(OINARI_FILES) bin/seed src/colonio.d.ts src/colonio_go.d.ts
	GOOS=js GOARCH=wasm go build -o ./dist/oinari.wasm ./cmd/agent/*.go
	GOOS=js GOARCH=wasm go build -o ./dist/sample.wasm ./cmd/sample/*.go
	npm run build

.PHONY: test-wasm
test-wasm: $(MAP_FILES)
	GOOS=js GOARCH=wasm go test -o ./dist/test_crosslink.wasm -c ./agent/crosslink/*
	GOOS=js GOARCH=wasm go test -o ./dist/test.wasm -c ./cmd/agent/*

test-ts: $(COLONIO_FILES) $(GO_FILES) $(OINARI_FILES) bin/seed src/keys.ts
	npm t

bin/seed: $(GO_FILES)
	go build -o $@ ./cmd/seed

dist/404.html: src/404.html keys.json
	go run ./cmd/tool template -i src/404.html -v keys.json > $@
  
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

dist/error.html: src/error.html keys.json
	go run ./cmd/tool template -i src/error.html -v keys.json > $@

dist/index.html: src/index.html keys.json
	go run ./cmd/tool template -i src/index.html -v keys.json > $@

dist/wasm_exec.js: $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	cp $< $@

src/colonio.d.ts: build/colonio/src/js/core.d.ts
	cp $< $@

src/colonio_go.d.ts: build/colonio/src/js/colonio_go.d.ts
	cp $< $@

src/keys.ts: src/keys.temp keys.json
	go run ./cmd/tool template -i src/keys.temp -v keys.json > $@
