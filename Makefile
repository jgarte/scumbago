BIN_DIR=build
BIN=$(BIN_DIR)/scumbago
BUILD_SHA=$(shell git describe --always --long --dirty)

BUILD_OPTS=-v -buildmode exe -o ${BIN}
LDFLAGS=-ldflags "-X github.com/Oshuma/scumbago/scumbag.BuildTag=${BUILD_SHA}"

.PHONY: all
all:
	[ -d ${BIN_DIR} ] || mkdir -p ${BIN_DIR}
	go build ${BUILD_OPTS} ${LDFLAGS}

.PHONY: clean
clean:
	if [ -f ${BIN} ] ; then rm -v ${BIN} ; fi
