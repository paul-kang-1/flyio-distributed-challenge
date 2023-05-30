#!/bin/bash

MAELSTROM_BIN="../maelstrom/maelstrom"
if [ ! -f $MAELSTROM_BIN ]; then
	echo "Maelstrom executable does not exist in ${MAELSTROM_BIN}."
	exit 1
fi

go get github.com/jepsen-io/maelstrom/demo/go
go install .

$MAELSTROM_BIN test \
	-w echo \
	--bin ~/go/bin/maelstrom-echo \
	--node-count 1 \
	--time-limit 10
