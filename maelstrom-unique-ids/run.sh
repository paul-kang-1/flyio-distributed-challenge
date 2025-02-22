#!/bin/bash

MAELSTROM_BIN="../maelstrom/maelstrom"
if [ ! -f $MAELSTROM_BIN ]; then
	echo "Maelstrom executable does not exist in ${MAELSTROM_BIN}."
	exit 1
fi

go get github.com/jepsen-io/maelstrom/demo/go
go install .

$MAELSTROM_BIN test \
	-w unique-ids \
	--bin ~/go/bin/maelstrom-unique-ids \
	--time-limit 30 \
	--rate 1000 \
	--node-count 3 \
	--availability total \
	--nemesis partition
