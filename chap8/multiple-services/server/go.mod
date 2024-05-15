module github.com/PaulOh5/multiple-services/server

go 1.22.2

require (
	github.com/PaulOh5/multiple-services/service v0.0.0
	google.golang.org/grpc v1.64.0
)

require (
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240513163218-0867130af1f8 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/PaulOh5/multiple-services/service => ../service
