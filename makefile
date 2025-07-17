.PHONY: build clean help

all: clean build

build:
	-go build -ldflags=-checklinkname=0 -v -o KittenCore main.go

clean:
	-rm KittenCore
	-go clean -i .
	-rm data/ai/user.yaml
	-rm data/zbp/banwords.yaml
	-rm data/Stack2/tips.yaml

help:
	@echo "make：编译包及其依赖"
	@echo "make clean：移除编译及缓存文件"