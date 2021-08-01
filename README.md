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

## TODO

- Finish the basic static website (show overall info, and point me to the APIs)
- Add goroutine that removes old protod instances (with some sort of timestamp)
- Merge all the tools into a single binary
- Return multiple versions of each config on the API
- Let the user configure Prototype, and revisit all the config params for the rest of the tools.
- Create Docker pipeline to create the image.
- Create example using docker compose.
- Add support for the rest of xDS configs.
- Add support for "original/static" envoy configs.
- Add support to apply a config to all the services with the same tag?
  - Maybe we can merge this config with the existing ones? That would be cool. Merging resources. Merge command?