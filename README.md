# Prototype

> ⚠️ Don't run this on production! Please.

Prototype is an open-source, easy-to-use, and minimal service mesh built on top of Envoy.

Compared to other service meshes, Prototype:
- does not create abstractions on top of Envoy.
- is universal. It can be used on Kubernetes, but it is not the main focus of the project.
- can be deployed and operated easily.
- is a **toy** project.

You can get a taste of Prototype, by running this [example](/example) scenario.


## Diagrams

![architecture](/media/architecture.png)
*Architecture*

![protod-detail](/media/protod-detail.png)  
*Protod detail*

## Inspiration

There are two main sources of inspiration for Prototype: [crossover](https://github.com/mumoshu/crossover) and [Kuma](https://kuma.io/).

- (some) Things I like from Crossover: 
  - The project doesn't try to make leaky abstractions on top of Envoy. 
  - It's simple and clean.

- (some) Things I like from Kuma: 
  - It's not solely focused on Kubernetes (as most meshes are). 
