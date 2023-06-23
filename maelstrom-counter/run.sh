#!/bin/bash

MAELSTROM_BIN="../maelstrom/maelstrom"
if [ ! -f $MAELSTROM_BIN ]; then
	echo "Maelstrom executable does not exist in ${MAELSTROM_BIN}."
	exit 1
fi

echo "💚 Building..."
go get github.com/jepsen-io/maelstrom/demo/go
go install .

echo "💚 Running"
$MAELSTROM_BIN test \
	-w g-counter \
	--bin ~/go/bin/maelstrom-counter \
	--node-count 3 \
	--rate 100 \
	--time-limit 20 \
	--nemesis partition
