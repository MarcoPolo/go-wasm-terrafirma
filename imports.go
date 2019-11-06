package terrafirma

// #include <stdlib.h>
//
// extern void hostcall_init_mm(void *context, int32_t x, int32_t y);
// extern void hostcall_panic_hook(void *context, int32_t x, int32_t y);
// extern int32_t hostcall_resp_set_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern int32_t hostcall_resp_set_body(void *context, int32_t a, int32_t b, int32_t c);
// extern int32_t hostcall_resp_set_response_code(void *context, int32_t a, int32_t b);
// extern int32_t hostcall_req_create(void *context, int32_t a, int32_t b, int32_t c, int32_t d);
// extern int32_t hostcall_req_set_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern int32_t hostcall_req_set_body (void *context, int32_t a, int32_t b, int32_t c);
// extern int32_t hostcall_req_send (void *context, int32_t a);
// extern int32_t hostcall_resp_get_response_code (void *context, int32_t a);
// extern void hostcall_resp_get_headers(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_resp_get_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern void hostcall_req_get_header(void *context, int32_t a, int32_t b, int32_t c, int32_t d, int32_t e);
// extern void hostcall_req_get_headers(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_method(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_body(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_resp_get_body(void *context, int32_t a, int32_t b, int32_t c);
// extern void hostcall_req_get_path(void *context, int32_t a, int32_t b, int32_t c);
import "C"

import (
	"context"
	"bytes"
	"unsafe"
	"strings"
	"net/http"
	ctxHttp "golang.org/x/net/context/ctxhttp"
	"fmt"
	"io/ioutil"
	"encoding/binary"
	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

const HostCallStatusOk = 0
const HostCallStatusInvalid = 1

// TODO use u32 when looking at memory

//export hostcall_init_mm
func hostcall_init_mm(context unsafe.Pointer, x int32, y int32) {
	// fmt.Println("init_mm")
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
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	body := memory.Data()[body_ptr:body_ptr+body_len]
	fmt.Println("Writing body", string(body), body_ptr, body_len)
	reqRespWrapper.ResponseWriter[resp].Write(body)

	return 0
}

//export hostcall_resp_set_response_code
func hostcall_resp_set_response_code(context unsafe.Pointer, resp int32, code int32) int32 {
	fmt.Println("here hostcall_resp_set_response_code", code, int(code))
	instanceContext := wasm.IntoInstanceContext(context);
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	reqRespWrapper.ResponseWriter[resp].WriteHeader(int(code))

	return 0
}

func get_header(memory *wasm.Memory, malloc func(...interface {}) (wasm.Value, error), header http.Header, values_ptr_p, values_len_p, req, name_ptr, name_len int32) {
	headerKey := string(memory.Data()[name_ptr:name_ptr+name_len])
	fmt.Println("Looking up header:", headerKey)
	headerVal := []byte(header.Get(headerKey))
	ptr, err := malloc(len(headerVal))
	if err != nil {
		panic("Failed to malloc")
	}
	fmt.Println("Header Val is", string(headerVal))

	var guestSlices []GuestSliceU8

	ptrI32 := ptr.ToI32()
	copy(memory.Data()[ptrI32:ptrI32+int32(len(headerVal))], headerVal)
	guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptrI32, len: int32(len(headerVal)) })

	writeGuestSlices(memory, malloc, guestSlices, values_ptr_p, values_len_p)

}

//export hostcall_req_get_header
func hostcall_req_get_header(context unsafe.Pointer, values_ptr_p, values_len_p, req, name_ptr, name_len int32) {
	fmt.Println("here hostcall_req_get_header")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]

	get_header(memory, reqRespWrapper.Malloc, reqRespWrapper.Request[req].Header, values_ptr_p, values_len_p, req, name_ptr, name_len)


	// headerKey := string(memory.Data()[name_ptr:name_ptr+name_len])
	// fmt.Println("Looking up header:", headerKey)
	// headerVal := []byte(reqRespWrapper.Request[req].Header.Get(headerKey))
	// ptr, err := reqRespWrapper.Malloc(len(headerVal))
	// if err != nil {
	// 	panic("Failed to malloc")
	// }
	// fmt.Println("Header Val is", string(headerVal))

	// var guestSlices []GuestSliceU8

	// ptrI32 := ptr.ToI32()
	// copy(memory.Data()[ptrI32:ptrI32+int32(len(headerVal))], headerVal)
	// guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptrI32, len: int32(len(headerVal)) })

	// writeGuestSlices(memory, reqRespWrapper.Malloc, guestSlices, values_ptr_p, values_len_p)
}

//export hostcall_resp_get_header
func hostcall_resp_get_header(context unsafe.Pointer, values_ptr_p, values_len_p, resp, name_ptr, name_len int32) {
	fmt.Println("here hostcall_resp_get_header")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]

	get_header(memory, reqRespWrapper.Malloc, reqRespWrapper.Response[resp].Header, values_ptr_p, values_len_p, resp, name_ptr, name_len)
}

func writeBytes(memory *wasm.Memory, malloc func(...interface {}) (wasm.Value, error), bytesToWrite []byte, value_ptr_p int32, value_len_p int32) {
	ptr, err := malloc(len(bytesToWrite))
	if err != nil {
		panic("Failed to malloc")
	}
	ptrI32 := ptr.ToI32()
	copy(memory.Data()[ptrI32:ptrI32 + int32(len(bytesToWrite))], bytesToWrite)

	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(ptrI32))
	copy(memory.Data()[value_ptr_p:value_ptr_p+4], bs)

	binary.LittleEndian.PutUint32(bs, uint32(len(bytesToWrite)))
	copy(memory.Data()[value_len_p:value_len_p+4], bs)
}

func readGuestSlicesAsString(memory *wasm.Memory, values_slice_ptr, values_slice_len int32) []string {
	output := []string{}
	fmt.Println("Values", values_slice_len)
	for i := int32(0); i < values_slice_len; i++ {
		start := values_slice_ptr + i*8
		var slice_start int32
		var slice_len int32
		buf := bytes.NewReader(memory.Data()[start:start+4])
		binary.Read(buf, binary.LittleEndian, &slice_start)
		buf = bytes.NewReader(memory.Data()[start+4:start+8])
		fmt.Println("Slice start", slice_start, slice_len)
		binary.Read(buf, binary.LittleEndian, &slice_len)
		s := string(memory.Data()[slice_start:slice_start+slice_len])
		output = append(output, s)
	}
	return output
}

func writeGuestSlices(memory *wasm.Memory, malloc func(...interface {}) (wasm.Value, error),  guestSlices []GuestSliceU8, values_ptr_p int32, values_len_p int32) {
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
		copy(memory.Data()[start:start+4], bs)
    binary.LittleEndian.PutUint32(bs, uint32(slice.len))
		copy(memory.Data()[start+4:start+8], bs)
	}

	binary.LittleEndian.PutUint32(bs, uint32(ptr_i32))
	copy(memory.Data()[values_ptr_p:values_ptr_p+4], bs)
	binary.LittleEndian.PutUint32(bs, uint32(sliceCount))
	copy(memory.Data()[values_len_p:values_len_p+4], bs)
}

// Maybe unused
func copyToGuest(memory *wasm.Memory, src []byte, ptr_i32, length int32) {
		if (ptr_i32 > int32(len(memory.Data()))) {
			memory.Grow(1)
			fmt.Println("Growing!!")
		}
		copy(memory.Data()[ptr_i32:ptr_i32+length], src)
}

func get_headers(memory *wasm.Memory, malloc func(...interface {}) (wasm.Value, error), header http.Header, headers_ptr_p, headers_len_p int32) {
	var guestSlices []GuestSliceU8
	for k := range header {
		ptr, err := malloc(len(k))
		if err != nil {
			panic("Failed to malloc")
		}
		copy(memory.Data()[ptr.ToI32():ptr.ToI32()+int32(len(k))], k)
		guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptr.ToI32(), len: int32(len(k)) })
	}

	writeGuestSlices(memory, malloc, guestSlices, headers_ptr_p, headers_len_p)
}

//export hostcall_req_get_headers
func hostcall_req_get_headers(context unsafe.Pointer, headers_ptr_p, headers_len_p, req int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	get_headers(memory, reqRespWrapper.Malloc, reqRespWrapper.Request[req].Header, headers_ptr_p, headers_len_p)
	// var guestSlices []GuestSliceU8
	// for k := range reqRespWrapper.Request[req].Header {
	// 	ptr, err := reqRespWrapper.Malloc(len(k))
	// 	if err != nil {
	// 		panic("Failed to malloc")
	// 	}
	// 	copy(memory.Data()[ptr.ToI32():ptr.ToI32()+int32(len(k))], k)
	// 	guestSlices = append(guestSlices, GuestSliceU8 { ptr: ptr.ToI32(), len: int32(len(k)) })
	// }

	// writeGuestSlices(memory, reqRespWrapper.Malloc, guestSlices, headers_ptr_p, headers_len_p)

	// fmt.Println("Guest slices", guestSlices)
	fmt.Println("Here in req_get_headers", reqRespWrapper.Request, req)
}

//export hostcall_resp_get_headers
func hostcall_resp_get_headers(context unsafe.Pointer, headers_ptr_p, headers_len_p, resp int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	fmt.Println("resp_get_headers")
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	get_headers(memory, reqRespWrapper.Malloc, reqRespWrapper.Response[resp].Header, headers_ptr_p, headers_len_p)
}

//export hostcall_req_get_method
func hostcall_req_get_method(context unsafe.Pointer, method_ptr_p, method_len_p, req int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	method := []byte(reqRespWrapper.Request[req].Method)
	fmt.Println("Here in req_get_method", string(method))
	writeBytes(memory, reqRespWrapper.Malloc, method, method_ptr_p, method_len_p)
}

//export hostcall_req_get_body
func hostcall_req_get_body(context unsafe.Pointer, body_ptr_p, body_len_p, req int32) {
	fmt.Println("hostcall_req_get_body")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	if reqRespWrapper.Request[req].Body == nil {
		fmt.Println("Nil body")
		return
	}

	body, err := ioutil.ReadAll(reqRespWrapper.Request[req].Body)
	if err != nil {
		fmt.Println("Err reading body", err)
		return
	}
	writeBytes(memory, reqRespWrapper.Malloc, body, body_ptr_p, body_len_p)
}

//export hostcall_resp_get_body
func hostcall_resp_get_body(context unsafe.Pointer, body_ptr_p, body_len_p, req int32) {
	fmt.Println("hostcall_req_get_body")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	if reqRespWrapper.Response[req].Body == nil {
		fmt.Println("Nil body")
		return
	}

	body, err := ioutil.ReadAll(reqRespWrapper.Response[req].Body)
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
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	url := []byte(reqRespWrapper.Request[req].URL.String())
	writeBytes(memory, reqRespWrapper.Malloc, url, path_ptr, path_len)
}

//export hostcall_panic_hook
func hostcall_panic_hook(context unsafe.Pointer, msg_ptr int32, msg_len int32) {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	fmt.Printf("WASM panicked: %v\n", string(memory.Data()[msg_ptr:msg_ptr+msg_len]))
	// TODO stop execution of wasm environment
	panic("Wasm panicked")
}

//export hostcall_req_set_header
func hostcall_req_set_header(
	context unsafe.Pointer,
	req,
	name_ptr,
	name_len,
	values_slice_ptr,
	values_slice_len int32,
) int32 {
	fmt.Println("req_set_header")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	request := reqRespWrapper.Request[req]
	headerKey := memory.Data()[name_ptr:name_ptr+name_len]
	fmt.Println("here")
	headerVals := readGuestSlicesAsString(memory, values_slice_ptr, values_slice_len)
	fmt.Println("Setting headers to", headerVals)
	request.Header.Set(string(headerKey), strings.Join(headerVals, ";"))
	return 0
}
//export hostcall_req_set_body
func hostcall_req_set_body(
	context unsafe.Pointer,
	req,
	body_ptr,
	body_len int32,
) int32 {
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	fmt.Println("req_set_body", *(*int)(instanceContext.Data()), req, reqRespWrapper.Request)
	request := reqRespWrapper.Request[req]
	body := memory.Data()[body_ptr:body_ptr+body_len]
	buf := bytes.NewReader(body)
	request.Body = ioutil.NopCloser(buf)
	return 0
}

//export hostcall_req_create
func hostcall_req_create(
	context unsafe.Pointer,
	method_ptr,
	method_len,
	url_ptr,
	url_len int32,
) int32 {
	fmt.Println("hostcall_req_create")
	instanceContext := wasm.IntoInstanceContext(context);
	memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	reqRespWrapper.Mutex.Lock()
	defer reqRespWrapper.Mutex.Unlock()

	method := memory.Data()[method_ptr:method_ptr + method_len]
	url := memory.Data()[url_ptr:url_ptr + url_len]

	r, err := http.NewRequest(string(method), string(url), nil)
	if err != nil {
		fmt.Println("Error in new req", err)
	}
	reqRespWrapper.Request = append(reqRespWrapper.Request, r)


	fmt.Println("id is", *(*int)(instanceContext.Data()), reqRespWrapper.Request, reqRespWrapper.Request[1])
	// reqRespWrapper.Request = []*http.Request{}
	return int32(len(reqRespWrapper.Request) - 1)
}

//export hostcall_req_send
func hostcall_req_send(
	ctx unsafe.Pointer,
	req int32,
) int32 {
	fmt.Println("req send")
	instanceContext := wasm.IntoInstanceContext(ctx);
	// memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	reqRespWrapper.Mutex.Lock()
	defer reqRespWrapper.Mutex.Unlock()
	request := reqRespWrapper.Request[req]

	// TODO add timeout
	resp, err := ctxHttp.Do(context.Background(), nil, request)
	if err != nil {
		fmt.Println("error in send req", err)
		return -1
	}

	reqRespWrapper.Response = append(reqRespWrapper.Response, resp)
	return int32(len(reqRespWrapper.Response) - 1)
}

//export hostcall_resp_get_response_code
func hostcall_resp_get_response_code(
	ctx unsafe.Pointer,
	resp int32,
) int32 {
	fmt.Println("get_response")
	instanceContext := wasm.IntoInstanceContext(ctx);
	// memory := instanceContext.Memory()
	reqRespWrapper := ReqContextMap[*(*int)(instanceContext.Data())]
	response := reqRespWrapper.Response[resp]

	return int32(response.StatusCode)
}

// GetImports returns the wasm imports
func GetImports() *wasm.Imports {
	imports := wasm.NewImports()

	imports.Namespace("env")
	var err error

	// imports, err = imports.Append("sum", sum, C.sum)
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
	imports, err = imports.Append("hostcall_resp_get_headers", hostcall_resp_get_headers, C.hostcall_resp_get_headers)
	imports, err = imports.Append("hostcall_resp_get_header", hostcall_resp_get_header, C.hostcall_resp_get_header)
	imports, err = imports.Append("hostcall_req_get_method", hostcall_req_get_method, C.hostcall_req_get_method)
	imports, err = imports.Append("hostcall_req_get_body", hostcall_req_get_body, C.hostcall_req_get_body)
	imports, err = imports.Append("hostcall_resp_get_body", hostcall_resp_get_body, C.hostcall_resp_get_body)
	imports, err = imports.Append("hostcall_req_get_path", hostcall_req_get_path, C.hostcall_req_get_path)
	imports, err = imports.Append("hostcall_req_create", hostcall_req_create, C.hostcall_req_create)
	imports, err = imports.Append("hostcall_req_set_header", hostcall_req_set_header, C.hostcall_req_set_header)
	imports, err = imports.Append("hostcall_req_set_body", hostcall_req_set_body, C.hostcall_req_set_body)
	imports, err = imports.Append("hostcall_req_send", hostcall_req_send, C.hostcall_req_send)
	imports, err = imports.Append("hostcall_resp_get_response_code", hostcall_resp_get_response_code, C.hostcall_resp_get_response_code)


	if err != nil {
		fmt.Printf("Err! %v\n", err)
	}
	return imports
}