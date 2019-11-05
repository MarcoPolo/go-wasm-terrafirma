package main

// #include <stdlib.h>
//
// extern int32_t sum(void *context, int32_t x, int32_t y);
// import "C"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"net/http"
	"net/http/httptest"

	"unsafe"

	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

type Response1 struct {
	Page   int
	Fruits []string
}

type GuestSliceU8 struct {
	ptr int32
	len int32
}

type Response2 struct {
	Page  int
	PageS []string
}

type ReqRespWrapper struct {
	Request  *http.Request
	Response http.ResponseWriter
	Malloc   (func(...interface{}) (wasm.Value, error))
}

var ReqContextMap = map[int]ReqRespWrapper{}
var currentId = 0

func main() {
	res1D := &Response1{
		Page:   1,
		Fruits: []string{"apple", "peach", "pear"}}
	res1B, _ := json.Marshal(res1D)

	fmt.Println(string(res1B))
	// Reads the WebAssembly module as bytes.
	bytes, _ := wasm.ReadBytes("rust_wasm_simple.wasm")
	// bytes, _ := wasm.ReadBytes("untouched.wasm")
	// wasi
	// bytes, _ := wasm.ReadBytes("rust-wasm-simple.wasm")
	// bytes, _ := wasm.ReadBytes("./wapm_packages/torch2424/wasm-matrix@0.0.2/optimized.wasm")
	// bytes, _ := wasm.ReadBytes("../hello-wasm2/go.wasm")

	imports := GetImports()
	fmt.Printf("imports: %v\n", imports)

	// Instantiates the WebAssembly module.
	instance, err := wasm.NewInstanceWithImports(bytes, imports)

	if err != nil {
		fmt.Printf("Err in creating! %v\n", err)
	}
	defer instance.Close()

	testReq := httptest.NewRequest("GET", "https://marcopolo.io/wasm", strings.NewReader("my request"))
	testReq.Header.Add("foo", "bar")
	testReq.Header.Add("baz", "boo")
	recorder := httptest.NewRecorder()
	// TODO fill in
	// wrapper := ReqRespWrapper{}
	ReqContextMap[0] = ReqRespWrapper{
		// wrapper := ReqRespWrapper{
		Request:  testReq,
		Response: recorder,
		Malloc:   instance.Exports["default_malloc_impl"]}
	wrapper := currentId
	currentId++
	// wrapper := 42
	fmt.Printf("Here %v\n", wrapper)
	// p := unsafe.Pointer(&wrapper)
	p := unsafe.Pointer(&wrapper)
	// fmt.Printf("Here %v\n", (*ReqRespWrapper)(p))
	fmt.Printf("Setting pointer to %v\n", p)
	instance.SetContextData(p)
	fmt.Printf("Foo %v\n", ReqContextMap[0])

	// Gets the `sum` exported function from the WebAssembly instance.
	// sum2 := instance.Exports["sum2"]

	// Calls that exported function with Go standard values. The WebAssembly
	// types are inferred and values are casted automatically.
	// result, _ := sum2(5, 37)
	// result2, _ := instance.Exports["add"](5, 37)
	// result2, _ := instance.Exports["sum3"](5, 37)
	fmt.Printf("exports: %v %v\n", instance.Exports, instance.Exports["_start"])
	result, _ := instance.Exports["sum3"](10, 0)
	fmt.Printf("sum3: %v\n", result)
	_, err = instance.Exports["run"]()
	fmt.Printf("run: %v\n", err)
	recorder.Header().Set("foo", "AtEnd1, AtEnd2")
	// recorder.WriteHeader(405)

	resp := recorder.Result()
	fmt.Printf("Response %v\n", resp)
	fmt.Println("Response Code", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response body", string(body))

	// result, _ := instance.Exports["add"](10, 2)
	// fmt.Printf("add: %v\n", result)
	// result2, _ := instance.Exports["__start"]()
	// fmt.Printf("start: %v\n", result2)
	// result, _ := instance.Exports["multiply"](10, 2)

	// fmt.Println(result)  // 42!
	// fmt.Println(result2) // 42!
}
