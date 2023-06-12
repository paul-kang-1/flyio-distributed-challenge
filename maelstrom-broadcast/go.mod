module maelstrom-broadcast

go 1.20

require (
	github.com/jepsen-io/maelstrom/demo/go v0.0.0-20230516124010-52951329816e
	github.com/paul-kang-1/flyio-distributed-challenge/utils v0.0.0-20230612203805-3714dd4184c7
)

replace github.com/paul-kang-1/flyio-distributed-challenge/utils => ../utils
