SHELL := /bin/bash

# Wallets
# Kennedy: 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32
# Pavel: 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4
# Ceasar: 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76
# Baba: 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9
# Ed: 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0
# Miner1: 0xFef311483Cc040e1A89fb9bb469eeB8A70935EF8
# Miner2: 0xb8Ee4c7ac4ca3269fEc242780D7D960bd6272a61
#
# Run two miners
# make up
# make up2
#
# Wallet Stuff
# go run app/wallet/cli/main.go generate
#
# Sample calls
# curl -il -X GET http://localhost:8080/v1/sample
# curl -il -X GET http://localhost:9080/v1/node/sample
# curl -il -X GET http://localhost:8080/v1/cancel
#

# ==============================================================================
# Local support

run: 
	go run app/scratch/main.go
	
up:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7281 --web-public-host 0.0.0.0:8280 --web-private-host 0.0.0.0:9280 --state-beneficiary=miner2 --state-db-path zblock/miner2/ | go run app/tooling/logfmt/main.go

up3:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7282 --web-public-host 0.0.0.0:8281 --web-private-host 0.0.0.0:9281 --state-beneficiary=miner3 --state-db-path zblock/miner3/ | go run app/tooling/logfmt/main.go

down:
	kill -INT $(shell ps | grep "main -race" | grep -v grep | sed -n 1,1p | cut -c1-5)

down-ubuntu:
	kill -INT $(shell ps -x | grep "main -race" | sed -n 1,1p | cut -c3-7)


# ==============================================================================
# Transactions
# PUBKEYS:
# Kenedy: 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32
# Pavel: 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4
# Ceaser: 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76
# Ed: 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0
# Baba: 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9

load: 
	go run app/wallet/cli/main.go send -a kennedy -n 1 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/wallet/cli/main.go send -a pavel -n 1 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
	go run app/wallet/cli/main.go send -a kennedy -n 2 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 150
	go run app/wallet/cli/main.go send -a pavel -n 2 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 125
	go run app/wallet/cli/main.go send -a kennedy -n 3 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 200
	go run app/wallet/cli/main.go send -a pavel -n 3 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 250
# ==============================================================================
# Modules support

load2:
	go run app/wallet/cli/main.go send -a kennedy -n 4 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/wallet/cli/main.go send -a pavel -n 4 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75


load3:
	go run app/wallet/cli/main.go send -a kennedy -n 5 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100

load4:
	go run app/wallet/cli/main.go send -a kennedy -n 6 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/wallet/cli/main.go send -a pavel -n 5 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
	go run app/wallet/cli/main.go send -a pavel -n 6 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
	
deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -v ./...
	go mod tidy
	go mod vendor

# ==============================================================================
# Running tests within the local computer
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	CGO_ENABLED=0 go test -count=1 ./...
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...
	govulncheck ./...