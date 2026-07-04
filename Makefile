.PHONY: test bench run clean

test:
	go test ./... -v

bench:
	go test ./bench/... -bench=. -benchmem -run=^$ -cpuprofile=cpu.prof -memprofile=mem.prof

run:
	go run ./cmd/server

clean:
	rm -rf data/*.wal data/*.snap cpu.prof mem.prof
