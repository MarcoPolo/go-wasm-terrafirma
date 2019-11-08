# Terrafirma backend

Check out the [blog post](https://marcopolo.io/wasm)

## Hello World Tutorial

[https://marcopolo.io/code/terrafirma/#terrafirma](https://marcopolo.io/code/terrafirma/#terrafirma)

## Running a Local Dev Version

1. `cd bin && go build main.go && cp main $GOPATH/bin/local-terrafirma`
2. Run the local server: `local-terrafirma hello.wasm`
3. In a separate terminal: `curl localhost:8081`
