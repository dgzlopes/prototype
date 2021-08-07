# Docker-compose example

Wanna test Prototype? You're on the right place!

> Note: This example is based on Envoy's [front-proxy sandbox](https://www.envoyproxy.io/docs/envoy/latest/start/sandboxes/front_proxy.html).

-> TODO: expand the example docs

docker-compose build (--pull)
docker-compose up

# Front configurations
go run cmd/protoctl/main.go apply -c default -s front -t cds -f ./example/configs/front/cds.yaml
go run cmd/protoctl/main.go apply -c default -s front -t lds -f ./example/configs/front/lds.yaml

# Service configurations
go run cmd/protoctl/main.go apply -c default -s service -t cds -f ./example/configs/service/cds.yaml
go run cmd/protoctl/main.go apply -c default -s service -t lds -f ./example/configs/service/lds.yaml

curl -v localhost:8080/service/1