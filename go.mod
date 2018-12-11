module github.com/kwf2030/commons

require (
	github.com/gorilla/websocket v1.4.0
	go.etcd.io/bbolt v1.3.1-etcd.8
	golang.org/x/net v0.0.0-20181207154023-610586996380
	golang.org/x/sys v0.0.0-20181210030007-2a47403f2ae5 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace golang.org/x/net v0.0.0-20181207154023-610586996380 => github.com/golang/net v0.0.0-20181207154023-610586996380

replace golang.org/x/sys v0.0.0-20181210030007-2a47403f2ae5 => github.com/golang/sys v0.0.0-20181210030007-2a47403f2ae5

replace go.etcd.io/bbolt v1.3.1-etcd.8 => github.com/etcd-io/bbolt v1.3.1-etcd.8
