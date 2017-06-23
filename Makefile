BIN_DIR=build
BIN=$(BIN_DIR)/scumbag
BUILD_SHA=$(shell git describe --always --long --dirty)

BUILD_OPTS=-i -v -buildmode exe -o ${BIN}
LDFLAGS=-ldflags "-X github.com/Oshuma/scumbago/scumbag.BUILD=${BUILD_SHA}"

.PHONY: all
all:
	[ -d ${BIN_DIR} ] || mkdir -p ${BIN_DIR}
	go build ${BUILD_OPTS} ${LDFLAGS}

.PHONY: clean
clean:
	if [ -f ${BIN} ] ; then rm -v ${BIN} ; fi
