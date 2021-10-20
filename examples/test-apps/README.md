**Steps to deploy a Go application on kubernetes with helm:**

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




