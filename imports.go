package main

// #include <stdlib.h>
//
// extern int32_t sum(void *context, int32_t x, int32_t y);
// extern void hostcall_init_mm(void *context, int32_t x, int32_t y);
// extern void hostcall_panic_hook(void *context, int32_t x, int32_t y);
// extern int32_t hostcall_resp_set_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern int32_t hostcall_resp_set_body(void *context, int32_t a, int32_t b, int32_t c);
// extern int32_t hostcall_resp_set_response_code(void *context, int32_t a, int32_t b);
// extern void hostcall_req_get_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern void hostcall_req_get_headers(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_method(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_body(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_path(void *context, int32_t a, int32_t b, int32_t c);
import "C"

import (
	"unsafe"
	"fmt"
	"io/ioutil"
	"encoding/binary"
	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

const HostCallStatusOk = 0
const HostCallStatusInvalid = 1

// TODO use u32 when looking at memory

//export sum
func sum(context unsafe.Pointer, x int32, y int32) int32 {
	return x + y
}

//export hostcall_init_mm
func hostcall_init_mm(context unsafe.Pointer, x int32, y int32) {
	fmt.Println("init_mm")
	// Noop
}

//export hostcall_resp_set_header
func hostcall_resp_set_header(context unsafe.Pointer, a, b, c, d, e int32) int32 {
	instanceContext := wasm.IntoInstanceContext(context);
	reqRespWrapper := (*ReqRespWrapper)(instanceContext.Data())
	fmt.Printf("set header %v\n", reqRespWrapper)
	return 0
}

//export hostcall_resp_set_body
func hostcall_resp_set_body(context unsafe.Pointer, resp, body_ptr, body_len int32) int32 {
	fmt.Println("hostcall_resp_set_body")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	body := memory[body_ptr:body_ptr+body_len]
	fmt.Println("Writing body", string(body), body_ptr, body_len)
	reqRespWrapper.Response.Write(body)

	return 0
}

//export hostcall_resp_set_response_code
func hostcall_resp_set_response_code(context unsafe.Pointer, resp int32, code int32) int32 {
	fmt.Println("here hostcall_resp_set_response_code", code, int(code))
	instanceContext := wasm.IntoInstanceContext(context);
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	reqRespWrapper.Response.WriteHeader(int(code))

	return 0
}

//export hostcall_req_get_header
func hostcall_req_get_header(context unsafe.Pointer, values_ptr_p, values_len_p, req, name_ptr, name_len int32) {
	fmt.Println("here hostcall_req_get_header")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]

	// values_ptr_p: *mut *mut GuestSlice<u8>,
	// values_len_p: *mut usize,
	// req: i32,
	// name_ptr: *const u8,
	// name_len: usize,

	headerKey := string(memory[name_ptr:name_ptr+name_len])
	fmt.Println("Looking up header:", headerKey)
	headerVal := []byte(reqRespWrapper.Request.Header.Get(headerKey))
	ptr, err := reqRespWrapper.Malloc(len(headerVal))
	if err != nil {
		panic("Failed to malloc")
	}
	fmt.Println("Header Val is", string(headerVal))

	var guestSlices []GuestSliceU8

	ptrI32 := ptr.ToI32()
	copy(memory[ptrI32:ptrI32+int32(len(headerVal))], headerVal)
	guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptrI32, len: int32(len(headerVal)) })

	writeGuestSlices(memory, reqRespWrapper.Malloc, guestSlices, values_ptr_p, values_len_p)
}

func writeBytes(memory []byte, malloc func(...interface {}) (wasm.Value, error), bytesToWrite []byte, value_ptr_p int32, value_len_p int32) {
	ptr, err := malloc(len(bytesToWrite))
	if err != nil {
		panic("Failed to malloc")
	}
	ptrI32 := ptr.ToI32()
	copy(memory[ptrI32:ptrI32 + int32(len(bytesToWrite))], bytesToWrite)

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(ptrI32))
	copy(memory[value_ptr_p:value_ptr_p+4], bs)

	binary.LittleEndian.PutUint32(bs, uint32(len(bytesToWrite)))
	copy(memory[value_len_p:value_len_p+4], bs)
}

func writeGuestSlices(memory []byte, malloc func(...interface {}) (wasm.Value, error),  guestSlices []GuestSliceU8, values_ptr_p int32, values_len_p int32) {
	fmt.Println("Writing Guest Slice")
	bs := make([]byte, 4)

	// each GuestSliceU8 is ptr (i32) + len (i32) = 8 bytes
	sliceCount := len(guestSlices)
	ptr, err := malloc(sliceCount * 8)
	if err != nil {
		panic("Failed to malloc")
	}


	ptr_i32 := ptr.ToI32()
	for idx, slice := range guestSlices {
		start := ptr_i32 + int32(idx)*8
		bs := make([]byte, 4)
    binary.LittleEndian.PutUint32(bs, uint32(slice.ptr))
		copy(memory[start:start+4], bs)
    binary.LittleEndian.PutUint32(bs, uint32(slice.len))
		copy(memory[start+4:start+8], bs)
	}

	binary.LittleEndian.PutUint32(bs, uint32(ptr_i32))
	copy(memory[values_ptr_p:values_ptr_p+4], bs)
	binary.LittleEndian.PutUint32(bs, uint32(sliceCount))
	copy(memory[values_len_p:values_len_p+4], bs)
}

//export hostcall_req_get_headers
func hostcall_req_get_headers(context unsafe.Pointer, headers_ptr_p, headers_len_p, req int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	var guestSlices []GuestSliceU8
	for k := range reqRespWrapper.Request.Header {
		ptr, err := reqRespWrapper.Malloc(len(k))
		if err != nil {
			panic("Failed to malloc")
		}
		copy(memory[ptr.ToI32():ptr.ToI32()+int32(len(k))], k)
		// fmt.Println(memory[ptr.ToI32():ptr.ToI32()+int32(len(k))])
		guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptr.ToI32(), len: int32(len(k)) })
	}

	writeGuestSlices(memory, reqRespWrapper.Malloc, guestSlices, headers_ptr_p, headers_len_p)

	// fmt.Println("Guest slices", guestSlices)
	fmt.Println("Here in req_get_headers", reqRespWrapper.Request, memory[headers_ptr_p], memory[headers_len_p], req)
}

//export hostcall_req_get_method
func hostcall_req_get_method(context unsafe.Pointer, method_ptr_p, method_len_p, req int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	method := []byte(reqRespWrapper.Request.Method)
	fmt.Println("Here in req_get_method", string(method))
	writeBytes(memory, reqRespWrapper.Malloc, method, method_ptr_p, method_len_p)
}

//export hostcall_req_get_body
func hostcall_req_get_body(context unsafe.Pointer, body_ptr_p, body_len_p, req int32) {
	fmt.Println("hostcall_req_get_body")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	if reqRespWrapper.Request.Body == nil {
		fmt.Println("Nil body")
		return
	}

	body, err := ioutil.ReadAll(reqRespWrapper.Request.Body)
	if err != nil {
		fmt.Println("Err reading body", err)
		return
	}
	writeBytes(memory, reqRespWrapper.Malloc, body, body_ptr_p, body_len_p)
}

//export hostcall_req_get_path
func hostcall_req_get_path(context unsafe.Pointer, path_ptr, path_len, req int32) {
	fmt.Println("hostcall_req_get_path")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	url := []byte(reqRespWrapper.Request.URL.String())
	writeBytes(memory, reqRespWrapper.Malloc, url, path_ptr, path_len)
}

//export hostcall_panic_hook
func hostcall_panic_hook(context unsafe.Pointer, msg_ptr int32, msg_len int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory().Data()
	fmt.Printf("WASM panicked: %v\n", string(memory[msg_ptr:msg_ptr+msg_len]))
	// TODO stop execution of wasm environment
	panic("Wasm panicked")
}

// GetImports returns the wasm imports
func GetImports() *wasm.Imports {
	imports := wasm.NewImports()

	imports.Namespace("env")
	var err error

	imports, err = imports.Append("sum", sum, C.sum)
	if err != nil {
		fmt.Printf("Err! %v\n", err)
	}

	// imports, err = imports.Append("fd_prestat_get", fd_prestat_get2, C.fd_prestat_get2)
	// if err != nil {
	// 	fmt.Printf("Err! %v\n", err)
	// }

	imports, err = imports.Append("hostcall_init_mm", hostcall_init_mm, C.hostcall_init_mm)
	imports, err = imports.Append("hostcall_panic_hook", hostcall_panic_hook, C.hostcall_panic_hook)
	imports, err = imports.Append("hostcall_resp_set_header", hostcall_resp_set_header, C.hostcall_resp_set_header)
	imports, err = imports.Append("hostcall_resp_set_body", hostcall_resp_set_body, C.hostcall_resp_set_body)
	imports, err = imports.Append("hostcall_resp_set_response_code", hostcall_resp_set_response_code, C.hostcall_resp_set_response_code)
	imports, err = imports.Append("hostcall_req_get_header", hostcall_req_get_header, C.hostcall_req_get_header)
	imports, err = imports.Append("hostcall_req_get_headers", hostcall_req_get_headers, C.hostcall_req_get_headers)
	imports, err = imports.Append("hostcall_req_get_method", hostcall_req_get_method, C.hostcall_req_get_method)
	imports, err = imports.Append("hostcall_req_get_body", hostcall_req_get_body, C.hostcall_req_get_body)
	imports, err = imports.Append("hostcall_req_get_path", hostcall_req_get_path, C.hostcall_req_get_path)


	if err != nil {
		fmt.Printf("Err! %v\n", err)
	}
	return imports
}