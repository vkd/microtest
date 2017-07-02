# Microtest - tool for testing microservices

Run tested image, send request and check expect response.

1. `go get github.com/vkd/microtest`
1. `cd $GOPATH/src/github.com/vkd/microtest`
1. Build docker image `./build.sh`
1. Create config file `./microtests/01-base.yaml`: 
```
image: my_microservice_image

tests:
  - name: check create user
    request:
      method: POST
      url: /cabinet/new
      body: |
        {
          "username": "John",
          "password": "password"
        }
    expect:
      status: 201
      body: |
        {
          "result": "ok"
        }

```
5. Start tests `microtest ./microtests/`
