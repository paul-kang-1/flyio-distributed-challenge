#!/bin/bash

MAELSTROM_BIN="../maelstrom/maelstrom"
if [ ! -f $MAELSTROM_BIN ]; then
	echo "Maelstrom executable does not exist in ${MAELSTROM_BIN}."
	exit 1
fi

part_a() {
	echo "💚 Running part a"
	$MAELSTROM_BIN test \
		-w broadcast \
		--bin ~/go/bin/maelstrom-broadcast \
		--node-count 1 \
		--time-limit 20 \
		--rate 10
}

part_b () {
	echo "💚 Running part b"
	$MAELSTROM_BIN test \
		-w broadcast \
		--bin ~/go/bin/maelstrom-broadcast \
		--node-count 5 \
		--time-limit 20 \
		--rate 10
}

part_c () {
	echo "💚 Running part c"
	$MAELSTROM_BIN test \
		-w broadcast \
		--bin ~/go/bin/maelstrom-broadcast \
		--node-count 5 \
		--time-limit 20 \
		--rate 10 \
		--nemesis partition
}


part_d () {
	echo "💚 Running part d"
	$MAELSTROM_BIN test \
		-w broadcast \
		--bin ~/go/bin/maelstrom-broadcast \
		--node-count 25 \
		--time-limit 20 \
		--rate 100 \
		--latency 100
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
		b) build; part_b ;;
		c) build; part_c ;;
		d) build; part_d ;;
		*) echo "💔 invalid part number" >&2; exit 1 ;;
	esac
}

main "$@"
