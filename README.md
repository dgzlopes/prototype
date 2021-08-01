# Prototype

Prototype is a **minimal** service mesh built on top of Envoy.

## Features

- **Minimal**
- **Envoy-based and compatible**: No need to change your Envoy configurations.
- **Universal:** Can run and operate anywhere.
## Architecture

![architecture](/media/architecture.png)

## Quickstart
Run Prototype:
```
go run cmd/prototype/prototype.go
```

Run a ProtoD instance:
```
go run cmd/protod/protod.go -cluster default -service quote -tags=env:production,version:0.0.6-beta
```

Apply some configs:
```
go run cmd/protoctl/main.go apply -c default -s quote -t cds -f ./example/configs/cds.yaml
go run cmd/protoctl/main.go apply -c default -s quote -t lds -f ./example/configs/lds.yaml
```