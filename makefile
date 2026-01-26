.PHONY: build clean help

all: clean build

build:
	@echo "复制 main.go 到内嵌资源……"
	mkdir -p kitten/data_internal/zbp
	cp main.go internal/config/prio/main.prio
	@echo "编译 KittenCore……"
	go build -ldflags="-s -w" -v -o KittenCore -trimpath

clean:
	-rm -f KittenCore
	-go clean
	-rm -f data/ai/user.yaml
	-rm -f data/zbp/banwords.yaml
	-rm -f data/Stack2/tips.yaml

help:
	@echo "make：编译包及其依赖"
	@echo "make clean：移除编译及缓存文件"