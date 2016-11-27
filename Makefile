SHORT_NAME := linkerd-proxy
BINDIR := ./bin

VERSION ?= git-$(shell git rev-parse --short HEAD)

REPO_PATH := github.com/ortoo/${SHORT_NAME}

IMAGE_NAME := quay.io/ortoo/${SHORT_NAME}

DEV_ENV_IMAGE := quay.io/ortoo/go
DEV_ENV_WORK_DIR := /go/src/${REPO_PATH}
DEV_ENV_CMD := docker run --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR} ${DEV_ENV_IMAGE}
DEV_ENV_CMD_INT := docker run -it --rm -v ${CURDIR}:${DEV_ENV_WORK_DIR} -w ${DEV_ENV_WORK_DIR} ${DEV_ENV_IMAGE}

dev:
	${DEV_ENV_CMD_INT} sh

# Containerized dependency resolution
bootstrap:
	${DEV_ENV_CMD} glide install

build:
	mkdir -p ${BINDIR}
	${DEV_ENV_CMD} make binary-build

binary-build:
	go build -o ${BINDIR}/${SHORT_NAME} ${SHORT_NAME}.go

build-docker: build
	docker build --rm -t ${IMAGE_NAME} .
	docker tag ${IMAGE_NAME} ${IMAGE_NAME}:${VERSION}
