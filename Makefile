
bin:
	go build -o i2p-stats main.go

fmt:
	gofumpt -w -s -extra .