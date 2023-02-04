
COLONIO_BRANCH := main
COLONIO_FILES := dist/colonio.js dist/colonio_go.js dist/colonio.wasm
GO_FILES := $(shell find . -name *.go)
OINARI_FILES := dist/wasm_exec.js

run: build
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

build: $(COLONIO_FILES) $(GO_FILES) $(OINARI_FILES) bin/seed dist/index.html src/keys.ts src/colonio.d.ts src/colonio_go.d.ts
	GOOS=js GOARCH=wasm go build -o ./dist/oinari.wasm ./cmd/agent/*.go
	GOOS=js GOARCH=wasm go test -o ./dist/test_crosslink.wasm -c ./agent/crosslink/*
	npm run build

test: $(COLONIO_FILES) $(GO_FILES) $(OINARI_FILES) bin/seed src/keys.ts
	npm t

bin/seed: $(GO_FILES)
	go build -o $@ ./cmd/seed

dist/colonio.js: build/colonio/output/colonio.js
	cp $< $@

dist/colonio.wasm: build/colonio/output/colonio.wasm
	cp $< $@

dist/colonio_go.js: build/colonio/src/js/colonio_go.js
	cp $< $@

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