# Cliente gRPC 
sera el encargado de enviar las peticiones al gRPC server que escribira los datos en el pub de rabbit mq y kafka


### generar archivos apartir de protoc
protoc --go_out=. --go-grpc_out=. protofiles/weather.proto
