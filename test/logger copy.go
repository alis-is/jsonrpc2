package test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"strings"
// )

// type req struct {
// 	Version string           `json:"jsonrpc"`
// 	Id      interface{}      `json:"id,omitempty"`
// 	Method  *string          `json:"method"`
// 	Params  *json.RawMessage `json:"params"`
// 	Result  *json.RawMessage `json:"result"`
// 	Error   *json.RawMessage `json:"error"`
// }

// type res struct {
// 	Id interface{} `json:"id"`
// }
// type e struct {
// 	Id string `json:"id"`
// }
// type f struct {
// 	Id string `json:"id"`
// }

// type a struct {
// 	e
// 	f
// 	Test string `json:"test"`
// }

// type decoder struct {
// 	json.Decoder
// 	r io.Reader
// }

// func NewDecoder(r io.Reader) *decoder {
// 	return &decoder{
// 		Decoder: *json.NewDecoder(r),
// 		r:       r,
// 	}
// }

// func (d *decoder) Decode(v interface{}) error {
// 	var tmp map[string]interface{}
// 	// we let json decoder to read entire value
// 	d.Decoder.Decode(&tmp)
// 	// then we parse it to rpc object
// 	fmt.Println(tmp)
// 	return nil
// }

// func main() {
// 	test := "{ \"jsonrpc\": \"2.0\", \"method\": \"subtract\", \"params\": [42, 23], \"id\": 1}"
// 	result := req{}
// 	_ = json.Unmarshal([]byte(test), &result)
// 	fmt.Println(string(*result.Params))

// 	test2 := req{
// 		Version: "2.0",
// 		Id:      nil,
// 	}
// 	fmt.Println(json.Marshal(test2))

// 	test3 := res{
// 		Id: nil,
// 	}
// 	r3, _ := json.Marshal(test3)
// 	fmt.Println(string(r3))
// 	test4 := res{
// 		Id: 1,
// 	}
// 	r4, _ := json.Marshal(test4)
// 	fmt.Println(string(r4))

// 	var test5 interface{} = "aaa"
// 	fmt.Println(test5 == "aaa")

// 	test6, _ := json.Marshal(a{e{"15"}, f{"16"}, "test"})
// 	fmt.Println("merged", string(test6))

// 	test7 := make([]interface{}, 0, 10)
// 	fmt.Println(len(test7))
// 	test7 = append(test7, 1)
// 	fmt.Println(len(test7))

// 	r := strings.NewReader("{ \"jsonrpc\": \"2.0\", \"method\": \"subtract\", \"params\": [42, 23], \"id\": 1}")
// 	d := NewDecoder(r)
// 	test8 := make(map[string]interface{})
// 	d.Decode(&test8)

// 	var x chan string
// 	x <- "test"
// 	close(x)
// 	x <- "test"
// }
