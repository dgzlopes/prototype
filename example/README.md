# Docker-compose example

Wanna test Prototype? You're in the right place!

> Note: This example is based on Envoy's [front-proxy sandbox](https://www.envoyproxy.io/docs/envoy/latest/start/sandboxes/front_proxy.html).

Build and start the docker-compose setup:
```
docker-compose build (--pull)
docker-compose up -d
```

Apply the LDS and CDS configurations for the front-proxy:
```
go run ../cmd/protoctl/main.go apply -c default -s front -t cds -f ./configs/front/cds.yaml
go run ../cmd/protoctl/main.go apply -c default -s front -t lds -f ./configs/front/lds.yaml
```

Apply the LDS and CDS configurations for the sidecar Protod running alongside the services:
```
go run ../cmd/protoctl/main.go apply -c default -s service -t cds -f ./configs/service/cds.yaml
go run ../cmd/protoctl/main.go apply -c default -s service -t lds -f ./configs/service/lds.yaml
```

Wait a bit for the configurations to propagate.

Check that we can access service one:
```
curl -v localhost:8080/service/1
```

You should see something like this:
```
.
.
Hello from behind Envoy (service 1)! hostname: d2d167fe6577 resolvedhostname: 192.168.208.5
```

Nice! Now let's try to scale the setup:
```
docker-compose up -d --scale service1=3 --no-recreate
```

Wait a bit for the new services to fetch the configurations.

Now, if you run multiple times that same request that we used before, you should see that the requests are being load-balanced across the three services:
```
.
.
Hello from behind Envoy (service 1)! hostname: d2d167fe6577 resolvedhostname: 192.168.208.5
.
.
Hello from behind Envoy (service 1)! hostname: 46cbbb9e9fed resolvedhostname: 192.168.208.6
.
.
Hello from behind Envoy (service 1)! hostname: b9bc774d2dba resolvedhostname: 192.168.208.7
```

Nice! Finally, let's try to update the listeners of the front-proxy:
```
go run ../cmd/protoctl/main.go apply -c default -s front -t lds -f ./configs/front/lds-swap.yaml
```

On that new configuration, the front-proxy will now route all the requests on `/service/1`, to service 2.

Let's check that! Run:
```
$ curl -v localhost:8080/service/1
```

And... Yup. Profit :)
```

Hello from behind Envoy (service 2)! hostname: 1390f1d44a0e resolvedhostname: 192.168.208.3
```

Once you're finished, you can shutdown the whole setup with:
```
docker-compose down
```