package terrafirma

import (
	"fmt"

	"net/http"

	"sync"
	"unsafe"

	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

type GuestSliceU8 struct {
	ptr int32
	len int32
}
type ReqRespWrapper struct {
	Mutex          *sync.Mutex
	Request        []*http.Request
	ResponseWriter []http.ResponseWriter
	Response       []*http.Response
	Malloc         (func(...interface{}) (wasm.Value, error))
}

var ReqContextMap = map[int]*ReqRespWrapper{}
var currentId = 0

func WasmHandler(wasmBytes []byte, imports *wasm.Imports, w http.ResponseWriter, r *http.Request) {
	// Instantiates the WebAssembly module.
	instance, err := wasm.NewInstanceWithImports(wasmBytes, imports)
	if err != nil {
		fmt.Printf("Err in creating! %v\n", err)
	}
	defer instance.Close()

	ReqContextMap[currentId] = &ReqRespWrapper{
		Mutex:          &sync.Mutex{},
		Request:        []*http.Request{r},
		ResponseWriter: []http.ResponseWriter{w},
		Response:       []*http.Response{},
		Malloc:         instance.Exports["default_malloc_impl"]}
	wrapper := currentId
	currentId++

	instance.SetContextData(unsafe.Pointer(&wrapper))
	_, err = instance.Exports["run"]()
}
