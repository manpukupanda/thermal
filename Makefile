# バイナリ名
BINARY=thermal

# ソースディレクトリ
SRC_DIR=cmd

# go mod tidy
.PHONY: tidy
tidy:
	go mod tidy

SRC_DIR := cmd
SRC_FILES := $(wildcard $(SRC_DIR)/*.go)

# コンパイル
.PHONY: build
build:
	@mkdir -p bin
	go build -o bin/$(BINARY) ./cmd

# 実行
.PHONY: run
run: build
	./bin/$(BINARY)

# クリーンアップ
.PHONY: clean
clean:
	rm -rf bin/$(BINARY)