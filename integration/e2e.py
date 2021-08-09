import os
import time
import requests
import sys

try:
    # Start docker-compose setup
    c = os.system("docker-compose -f example/docker-compose.yaml up -d --remove-orphans")
    assert c != 0, "docker-compose up failed"

    # Wait a bit
    time.sleep(5)

    # Create initial configurations
    c += os.system("go run cmd/protoctl/main.go apply -c default -s front -t cds -f example/configs/front/cds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s front -t lds -f example/configs/front/lds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s service -t cds -f example/configs/service/cds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s service -t lds -f example/configs/service/lds.yaml")
    assert c == 0, "failed to create initial configurations"

    # Check if the control plane detected all the protods on the data plane
    r = requests.get("http://localhost:10000/api/protod")
    assert r.status_code == 200, "failed to fetch protods"
    assert len(r.json()) == 3, "failed to find all protods"

    # Wait a bit - so protods can fetch the new configurations
    time.sleep(10)

    # Check if the routing is working
    r = requests.get("http://localhost:8080/service/1")
    assert r.status_code == 200, "failed to call service 1"
    assert "Hello from behind Envoy (service 1)!" in r.text, "response incorrect from service 1"
    r = requests.get("http://localhost:8080/service/2")
    assert r.status_code == 200, "failed to call service 2"
    assert "Hello from behind Envoy (service 2)!" in r.text, "response incorrect from service 2"

    # Scale number of services
    c = os.system("docker-compose -f example/docker-compose.yaml up -d --scale service1=3 --no-recreate")
    assert c == 0, "docker-compose scale failed"

    # Wait a bit - so docker-compose can scale the setup
    time.sleep(10)

    # Check if load balancing is working
    ips = set()
    for x in range(5):
        r = requests.get("http://localhost:8080/service/1")
        assert r.status_code == 200, "failed to call service 1"
        assert "Hello from behind Envoy (service 1)!" in r.text, "response incorrect from service 1"
        ips.add(r.text.split(" ")[-1])
    assert len(ips) == 3, "load balacing isn't working properly"

    c = os.system("go run cmd/protoctl/main.go apply -c default -s front -t lds -f example/configs/front/lds-swap.yaml")
    assert c == 0, "failed update front lds config"

    # Wait a bit - so protods can fetch the new configurations
    time.sleep(10)

    # Check if new config was applied properly (we call service 1, but Envoy should route us to service 2)
    r = requests.get("http://localhost:8080/service/1")
    assert r.status_code == 200, "failed to call service 1"
    assert "Hello from behind Envoy (service 2)!" in r.text, "response incorrect from service 1"

    # Check (again) if the control plane detected all the protods on the data plane
    r = requests.get("http://localhost:10000/api/protod")
    assert r.status_code == 200, "failed to fetch protods"
    assert len(r.json()) == 5, "failed to find all protods"

except Exception as e:
    # Show exception
    print(e)
    # Shutdown docker-compose
    c = os.system("docker-compose -f example/docker-compose.yaml down")
    assert c == 0, "docker-compose down failed"
    sys.exit(os.EX_SOFTWARE)

else:
    # Shutdown docker-compose
    c = os.system("docker-compose -f example/docker-compose.yaml down")
    assert c == 0, "docker-compose down failed"
    sys.exit(os.EX_OK)