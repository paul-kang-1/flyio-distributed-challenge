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

main() {
	go get github.com/jepsen-io/maelstrom/demo/go
	go install .

	if [[ $# -ge 2 ]]; then echo "usage: $0 [<part>]" >&2; exit 1; fi

	if [[ $# -eq 0 ]]; then part_a; part_b; exit 0; fi

	case "$1" in
		a) part_a ;;
		b) part_b ;;
		d) part_d ;;
		*) echo "💔 invalid part number" >&2; exit 1 ;;
	esac
}

main "$@"
