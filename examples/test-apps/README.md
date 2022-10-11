**Steps to create server and client images **

(1) Compile the code under app-code
```
    cd examples/test-apps/http-server/
    go build http-server.go
    cd ../http-client/
    go build http-client.go
```

(2) Build the docker images
```
    cd ../
    docker build -t httptest-base-server -f Dockerfile_server .
    docker build -t httptest-base-client -f Dockerfile_client .
```

(3) Tag the images with docker registry
```
    docker tag httptest-base-server:latest <docker-registry-url>/my-custom-httptest-server:1.1
    docker tag httptest-base-client:latest <docker-registry-url>/my-custom-httptest-client:1.1
    Note: Bump up the version if you change the code
```

(4) Push these images to docker registry
```
    docker push <docker-registry-url>/my-custom-httptest-server:1.1
    docker push <docker-registry-url>/my-custom-httptest-client:1.1
```

(5) Modify the helm files (values.yaml, service.yaml and deployment.yaml) accordingly in the folder examples/helm_charts/http-client/helm/http-client and examples/helm_charts/http-server/helm/http-server
    Note: The NodePort in values.yaml is the port exposed by the service running on K8s. Also update the tag of the image to be downloaded if required


**Steps to create httpbin sleep with curl image **

(1) Build the docker images
```
    cd examples/test-apps/
    docker build -t httpbin-client -f Dockerfile_sleep .
```

(2) Tag the images with docker registry
```
    docker tag httpbin-client:latest <docker-registry-url>/httpbin-client:1.0
```

(3) Push these images to docker registry
```
    docker push <docker-registry-url>/httpbin-client:1.0
```
**Steps to build grpc xstream server client images:**
(1) Clone the repo
```
    cd examples/test-apps
    git clone https://github.com/toransahu/grpc-eg-go/
```
(2) Compile client and server
```
    cd grpc-eg-go/cmd;go build run_machine_server.go
    cd ../client;go build machine.go
```
(3) Build client and server images
```
    cd ../..
    docker build -t xstream-server -f Dockerfile_xstream-server .
    docker build -t xstream-client -f Dockerfile_xstream-client .
```
(4) Tag the images
```
    docker tag xstream-server:latest <docker-registry-url>/xstream-server:1.0
    docker tag xstream-client:latest <docker-registry-url>/xstream-client:1.0
```
(5) Push the images to docker repo
```
    docker push <docker-registry-url>/xstream-server:1.0
    docker push <docker-registry-url>/xstream-client:1.0
```
