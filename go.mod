module github.com/kwf2030/commons

require (
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23
	github.com/gorilla/websocket v1.4.0
	go.etcd.io/bbolt v1.3.1-etcd.8
	golang.org/x/net v0.0.0-20190110200230-915654e7eabc
	golang.org/x/sys v0.0.0-20190115152922-a457fd036447 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace golang.org/x/net v0.0.0-20190110200230-915654e7eabc => github.com/golang/net v0.0.0-20190110200230-915654e7eabc

replace golang.org/x/sys v0.0.0-20190115152922-a457fd036447 => github.com/golang/sys v0.0.0-20190115152922-a457fd036447

replace go.etcd.io/bbolt v1.3.1-etcd.8 => github.com/etcd-io/bbolt v1.3.1-etcd.8
