module github.com/sfi2k7/blueconfig

go 1.24.7

require (
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/sfi2k7/microweb v0.0.0-20251016174507-8e112fea3fc6
	go.etcd.io/bbolt v1.4.3
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace github.com/sfi2k7/microweb => ../microweb
