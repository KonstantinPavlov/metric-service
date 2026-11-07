# api

В этой директории принято размещать proto-файлы или файлы в формате OpenAPI/Swagger для описания контракта сервиса.

Protocol Buffers (Protobuf) будет изучаться дальше по курсу.


## Команды генерации
```bash 
mkdir -p internal/proto

protoc --go_out=internal/proto --go_opt=paths=source_relative \
       --go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative \
       -I api/proto api/proto/metrics.proto
```