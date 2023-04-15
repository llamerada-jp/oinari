
COLONIO_BRANCH := main
COLONIO_FILES := dist/colonio.js dist/colonio_go.js dist/colonio.wasm
GO_FILES := $(shell find . -name *.go | grep -v ./build/)
TS_FILES := $(shell find ./src/ -name *.ts) src/colonio.d.ts src/colonio_go.d.ts src/keys.ts
OINARI_FILES := dist/wasm_exec.js dist/404.html dist/error.html dist/index.html
MAP_FILES := dist/colonio.wasm.map dist/colonio_go.js.map

build: $(COLONIO_FILES) $(OINARI_FILES) build-go build-ts

build-go: $(GO_FILES) go.mod go.sum
	GOOS=js GOARCH=wasm go build -o ./dist/oinari.wasm ./cmd/node/*.go
	GOOS=js GOARCH=wasm go build -o ./dist/test/exit.wasm ./cmd/app/exit/*.go
	GOOS=js GOARCH=wasm go build -o ./dist/test/sleep.wasm ./cmd/app/sleep/*.go
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_crosslink.wasm -c ./lib/crosslink/
	## should edit TESTS@src/test.ts to run test build by wasm
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_api.wasm -c ./api/
	GOOS=js GOARCH=wasm go test -o ./dist/test/test_node.wasm -c ./cmd/node/

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
s: build
	while true; do go run ./cmd/test_seed; done

.PHONY: test
test: build
	npm t
	# go run ./cmd/test_seed -test

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

.PHONY: clean
clean:
	rm -f ./dist/**/*.js ./dist/**/*.map ./dist/**/*.wasm $(OINARI_FILES)