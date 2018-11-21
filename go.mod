module github.com/kwf2030/commons

require (
	github.com/gorilla/websocket v1.4.0
	github.com/rs/zerolog v1.10.3
	go.etcd.io/bbolt v1.3.1-etcd.8
	golang.org/x/net v0.0.0-20181114220301-adae6a3d119a
	golang.org/x/sys v0.0.0-20181107165924-66b7b1311ac8 // indirect
	gopkg.in/yaml.v2 v2.2.1
)

replace golang.org/x/net v0.0.0-20181114220301-adae6a3d119a => github.com/golang/net v0.0.0-20181114220301-adae6a3d119a

replace golang.org/x/sys v0.0.0-20181107165924-66b7b1311ac8 => github.com/golang/sys v0.0.0-20181107165924-66b7b1311ac8
