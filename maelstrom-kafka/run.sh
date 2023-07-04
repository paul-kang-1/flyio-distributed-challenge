#!/bin/bash

MAELSTROM_BIN="../maelstrom/maelstrom"
if [ ! -f $MAELSTROM_BIN ]; then
	echo "Maelstrom executable does not exist in ${MAELSTROM_BIN}."
	exit 1
fi

part_a() {
	echo "💚 Running part a"
	$MAELSTROM_BIN test \
		-w kafka \
		--bin ~/go/bin/maelstrom-kafka \
		--node-count 1 \
		--concurrency 2n \
		--time-limit 20 \
		--rate 1000
}

build () {
	echo "💚 Building..."
	go get github.com/jepsen-io/maelstrom/demo/go
	go install .
}

main() {
	if [[ $# -ne 1 ]]; then echo "usage: $0 [<part>]" >&2; exit 1; fi

	case "$1" in
		a) build; part_a ;;
		*) echo "💔 invalid part number" >&2; exit 1 ;;
	esac
}

main "$@"
