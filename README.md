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

## Inspiration

There are two main sources of inspiration for Prototype: [crossover](https://github.com/mumoshu/crossover) and [Kuma](https://kuma.io/).

- (some) Things I like from Crossover: 
  - The project doesn't try to make leaky abstractions op top of Envoy. 
  - It's simple and clean.

- (some) Things I like from Kuma: 
  - It's not solely focused on Kubernetes (as most meshes are). 
