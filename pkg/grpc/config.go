package grpc

type PublicGrpcConfig struct {
	GrpcProtocol string `env:"GRPC_PROTOCOL, default=tcp"`
	GrpcAddress  string `env:"GRPC_ADDRESS, default=:32023"`
}
