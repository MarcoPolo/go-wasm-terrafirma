package terrafirma

// #include <stdlib.h>
//
// extern int32_t sum(void *context, int32_t x, int32_t y);
// import "C"

import (
	"fmt"
	"io/ioutil"
	"strings"

	"net/http"
	"net/http/httptest"

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

const wasmTarget = "rust_wasm_simple.wasm"

func main() {
	// TestGet()
	TestPost()
}

func TestPost() {
	testReq := httptest.NewRequest("POST", "https://marcopolo.io/wasm", strings.NewReader("my request"))
	testReq.Header.Add("foo", "bar")
	testReq.Header.Add("baz", "boo")
	recorder := httptest.NewRecorder()

	bytes, _ := wasm.ReadBytes(wasmTarget)
	imports := GetImports()

	WasmHandler(bytes, imports, recorder, testReq)
	resp := recorder.Result()
	fmt.Printf("Response %v\n", resp)
	fmt.Println("Response Code", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response body", string(body))
}

func TestGet() {
	testReq := httptest.NewRequest("GET", "https://marcopolo.io/wasm", strings.NewReader("my request"))
	testReq.Header.Add("foo", "bar")
	testReq.Header.Add("baz", "boo")
	recorder := httptest.NewRecorder()

	bytes, _ := wasm.ReadBytes(wasmTarget)
	imports := GetImports()

	WasmHandler(bytes, imports, recorder, testReq)
	resp := recorder.Result()
	fmt.Printf("Response %v\n", resp)
	fmt.Println("Response Code", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response body", string(body))
}

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
