import os
import time
import requests
import sys

try:
    # Start docker-compose setup
    c = os.system("docker run -d --network host ghcr.io/dgzlopes/prototype:latest")
    assert c == 0, "failed to start prototype container"

    # Wait a bit
    time.sleep(10)

    # Create initial configurations
    c += os.system("go run cmd/protoctl/main.go apply -c default -s front -t cds -f example/configs/front/cds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s front -t lds -f example/configs/front/lds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s service -t cds -f example/configs/service/cds.yaml")
    c += os.system("go run cmd/protoctl/main.go apply -c default -s service -t lds -f example/configs/service/lds.yaml")
    assert c == 0, "failed to create initial configurations"

    # Run protod k6 test
    c = os.system("docker run -v $PWD:/src -i --network host loadimpact/k6 run --quiet /src/integration/protod.js")
    assert c == 0, "failed to run protod k6 test"

except Exception as e:
    # Show exception
    print(e)
    # Stop all the things
    c = os.system("docker stop $(docker ps -a -q)")
    assert c == 0, "failed to stop all the containers"
    sys.exit(os.EX_SOFTWARE)

else:
    # Stop all the things
    c = os.system("docker stop $(docker ps -a -q)")
    assert c == 0, "failed to stop all the containers"
    sys.exit(os.EX_OK)