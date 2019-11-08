package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	terrafirma "github.com/marcopolo/go-wasm-terrafirma"
	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage `terrafirma-bin")
	}
	wasmFile := os.Args[1]
	fmt.Println("Using", wasmFile)
	imports := terrafirma.GetImports()
	bytes, _ := wasm.ReadBytes(wasmFile)
	handler := terrafirma.NewWasmHandler(bytes, imports)

	http.Handle("/", handler)
	fmt.Println("Running on http://localhost:8081/")
	log.Fatal(http.ListenAndServe(":8081", nil))

}
