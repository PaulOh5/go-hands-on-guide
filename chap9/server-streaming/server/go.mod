module github.com/PaulOh5/server-streaming/server

go 1.22.2

require (
	github.com/PaulOh5/server-streaming/service v0.0.0
	google.golang.org/grpc v1.64.0
)

require (
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/PaulOh5/server-streaming/service => ../service
