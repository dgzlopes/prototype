# Prototype

> ⚠️ Don't run this on production! It's a proof-of-concept.

Prototype is an open source, easy-to-use and minimal service mesh built on top of Envoy.

## Quickstart

Run Prototype:
```
go run cmd/prototype/prototype.go
```

Run a ProtoD instance:
```
go run cmd/prototype/prototype.go -d -cluster default -service quote -tags=env:production,version:0.0.6-beta
```

Apply some configs:
```
go run cmd/protoctl/main.go apply -c default -s quote -t cds -f ./example/configs/cds.yaml
go run cmd/protoctl/main.go apply -c default -s quote -t lds -f ./example/configs/lds.yaml
```


## Diagrams

![architecture](/media/architecture.png)
*Architecture*

![protod-detail](/media/protod-detail.png)  
*Protod detail*


<details  style="margin-left:1.2em;">
    <summary><b>Internal diagrams</b></summary>

 
![internal-kv-datamodel](/media/internal-kv-datamodel.png)  
*Internal K/V datamodel*

![protod-internal-flows](/media/protod-internal-flows.png)  
*Internal Protod flows*
</details>

## Inspiration

There are two main sources of inspiration for Prototype: [crossover](https://github.com/mumoshu/crossover) and [Kuma](https://kuma.io/).

- Crossover a minimal and sufficient xDS for Envoy. 
  - The project doesn't try to make leaky abstractions op top of Envoy. 
  - It's simple and clean.

- Kuma is an universal Envoy service mesh. 
  - It's not solely focused on Kubernetes (as most meshes are). 
    - Not everyone has everything on Kubernetes :)

## TODO

- Finish the basic static website (show overall info, and point me to the APIs)
- ~~Add goroutine that removes old protod instances (with some sort of timestamp)~~ (we don't need this)
- ~~Merge all the tools into a single binary~~
- Return multiple versions of each config on the API
- Let the user configure Prototype, and revisit all the config params for the rest of the tools.
- ~~Create Docker pipeline to create the image.~~
- ~~Create example using docker compose.~~
  - Clean up the example, and add docs.
- Add support for the rest of xDS configs.
- Add support for "original/static" envoy configs.
- Re-check the things that we don't have flags for (hardcoded stuff)
- Add support to apply a config to all the services with the same tag?
  - Maybe we can merge this config with the existing ones? That would be cool. Merging resources. Merge command?
- Autodetect resource type from the file and be able to mix multiple resources on the same file (of diff types).

- [IF I'VE TIME] Add support for etcd! HA :) Remember the locks daniel
