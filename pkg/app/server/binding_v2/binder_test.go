package binding_v2

import (
	"fmt"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server/binding"
	"github.com/cloudwego/hertz/pkg/common/test/assert"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route/param"
)

type mockRequest struct {
	Req *protocol.Request
}

func newMockRequest() *mockRequest {
	return &mockRequest{
		Req: &protocol.Request{},
	}
}

func (m *mockRequest) SetRequestURI(uri string) *mockRequest {
	m.Req.SetRequestURI(uri)
	return m
}

func (m *mockRequest) SetHeader(key, value string) *mockRequest {
	m.Req.Header.Set(key, value)
	return m
}

func (m *mockRequest) SetHeaders(key, value string) *mockRequest {
	m.Req.Header.Set(key, value)
	return m
}

func (m *mockRequest) SetPostArg(key, value string) *mockRequest {
	m.Req.PostArgs().Add(key, value)
	return m
}

func (m *mockRequest) SetUrlEncodeContentType() *mockRequest {
	m.Req.Header.SetContentTypeBytes([]byte("application/x-www-form-urlencoded"))
	return m
}

func (m *mockRequest) SetJSONContentType() *mockRequest {
	m.Req.Header.SetContentTypeBytes([]byte(jsonContentTypeBytes))
	return m
}

func (m *mockRequest) SetBody(data []byte) *mockRequest {
	m.Req.SetBody(data)
	m.Req.Header.SetContentLength(len(data))
	return m
}

func TestBind_BaseType(t *testing.T) {
	bind := Bind{}
	type Req struct {
		Version int    `path:"v"`
		ID      int    `query:"id"`
		Header  string `header:"H"`
		Form    string `form:"f"`
	}

	req := newMockRequest().
		SetRequestURI("http://foobar.com?id=12").
		SetHeaders("H", "header").
		SetPostArg("f", "form").
		SetUrlEncodeContentType()
	var params param.Params
	params = append(params, param.Param{
		Key:   "v",
		Value: "1",
	})

	var result Req

	err := bind.Bind(req.Req, params, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, 1, result.Version)
	assert.DeepEqual(t, 12, result.ID)
	assert.DeepEqual(t, "header", result.Header)
	assert.DeepEqual(t, "form", result.Form)
}

func TestBind_SliceType(t *testing.T) {
	bind := Bind{}
	type Req struct {
		ID   []int     `query:"id"`
		Str  [3]string `query:"str"`
		Byte []byte    `query:"b"`
	}
	IDs := []int{11, 12, 13}
	Strs := [3]string{"qwe", "asd", "zxc"}
	Bytes := []byte("123")

	req := newMockRequest().
		SetRequestURI(fmt.Sprintf("http://foobar.com?id=%d&id=%d&id=%d&str=%s&str=%s&str=%s&b=%d&b=%d&b=%d", IDs[0], IDs[1], IDs[2], Strs[0], Strs[1], Strs[2], Bytes[0], Bytes[1], Bytes[2]))

	var result Req

	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, 3, len(result.ID))
	for idx, val := range IDs {
		assert.DeepEqual(t, val, result.ID[idx])
	}
	assert.DeepEqual(t, 3, len(result.Str))
	for idx, val := range Strs {
		assert.DeepEqual(t, val, result.Str[idx])
	}
	assert.DeepEqual(t, 3, len(result.Byte))
	for idx, val := range Bytes {
		assert.DeepEqual(t, val, result.Byte[idx])
	}
}

func TestBind_StructType(t *testing.T) {
	type FFF struct {
		F1 string `query:"F1"`
	}

	type TTT struct {
		T1 string `query:"F1"`
		T2 FFF
	}

	type Foo struct {
		F1 string `query:"F1"`
		F2 string `header:"f2"`
		F3 TTT
	}

	type Bar struct {
		B1 string `query:"B1"`
		B2 Foo    `query:"B2"`
	}

	bind := Bind{}

	var result Bar

	req := newMockRequest().SetRequestURI("http://foobar.com?F1=f1&B1=b1").SetHeader("f2", "f2")

	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, "b1", result.B1)
	assert.DeepEqual(t, "f1", result.B2.F1)
	assert.DeepEqual(t, "f2", result.B2.F2)
	assert.DeepEqual(t, "f1", result.B2.F3.T1)
	assert.DeepEqual(t, "f1", result.B2.F3.T2.F1)
}

func TestBind_PointerType(t *testing.T) {
	type TT struct {
		T1 string `query:"F1"`
	}

	type Foo struct {
		F1 *TT                       `query:"F1"`
		F2 *******************string `query:"F1"`
	}

	type Bar struct {
		B1 ***string `query:"B1"`
		B2 ****Foo   `query:"B2"`
		B3 []*string `query:"B3"`
		B4 [2]*int   `query:"B4"`
	}

	bind := Bind{}

	result := Bar{}

	F1 := "f1"
	B1 := "b1"
	B2 := "b2"
	B3s := []string{"b31", "b32"}
	B4s := [2]int{0, 1}

	req := newMockRequest().SetRequestURI(fmt.Sprintf("http://foobar.com?F1=%s&B1=%s&B2=%s&B3=%s&B3=%s&B4=%d&B4=%d", F1, B1, B2, B3s[0], B3s[1], B4s[0], B4s[1])).
		SetHeader("f2", "f2")

	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, B1, ***result.B1)
	assert.DeepEqual(t, F1, (*(****result.B2).F1).T1)
	assert.DeepEqual(t, F1, *******************(****result.B2).F2)
	assert.DeepEqual(t, len(B3s), len(result.B3))
	for idx, val := range B3s {
		assert.DeepEqual(t, val, *result.B3[idx])
	}
	assert.DeepEqual(t, len(B4s), len(result.B4))
	for idx, val := range B4s {
		assert.DeepEqual(t, val, *result.B4[idx])
	}
}

func TestBind_NestedStruct(t *testing.T) {
	type Foo struct {
		F1 string `query:"F1"`
	}

	type Bar struct {
		Foo
		Nested struct {
			N1 string `query:"F1"`
		}
	}

	bind := Bind{}

	result := Bar{}

	req := newMockRequest().SetRequestURI("http://foobar.com?F1=qwe")
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, "qwe", result.Foo.F1)
	assert.DeepEqual(t, "qwe", result.Nested.N1)
}

func TestBind_SliceStruct(t *testing.T) {
	type Foo struct {
		F1 string `json:"f1"`
	}

	type Bar struct {
		B1 []Foo `query:"F1"`
	}

	bind := Bind{}

	result := Bar{}
	B1s := []string{"1", "2", "3"}

	req := newMockRequest().SetRequestURI(fmt.Sprintf("http://foobar.com?F1={\"f1\":\"%s\"}&F1={\"f1\":\"%s\"}&F1={\"f1\":\"%s\"}", B1s[0], B1s[1], B1s[2]))
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, len(result.B1), len(B1s))
	for idx, val := range B1s {
		assert.DeepEqual(t, B1s[idx], val)
	}
}

func TestBind_MapType(t *testing.T) {
	var result map[string]string
	bind := Bind{}
	req := newMockRequest().
		SetJSONContentType().
		SetBody([]byte(`{"j1":"j1", "j2":"j2"}`))
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, 2, len(result))
	assert.DeepEqual(t, "j1", result["j1"])
	assert.DeepEqual(t, "j2", result["j2"])
}

func TestBind_MapFieldType(t *testing.T) {
	type Foo struct {
		F1 ***map[string]string `query:"f1" json:"f1"`
	}

	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?f1={\"f1\":\"f1\"}").
		SetJSONContentType().
		SetBody([]byte(`{"j1":"j1", "j2":"j2"}`))
	result := Foo{}
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, 1, len(***result.F1))
	assert.DeepEqual(t, "f1", (***result.F1)["f1"])
}

func TestBind_UnexportedField(t *testing.T) {
	var s struct {
		A int `query:"a"`
		b int `query:"b"`
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?a=1&b=2")
	err := bind.Bind(req.Req, nil, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, 1, s.A)
	assert.DeepEqual(t, 0, s.b)
}

func TestBind_NoTagField(t *testing.T) {
	var s struct {
		A string
		B string
		C string
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?B=b1&C=c1").
		SetHeader("A", "a2")

	var params param.Params
	params = append(params, param.Param{
		Key:   "B",
		Value: "b2",
	})

	err := bind.Bind(req.Req, params, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, "a2", s.A)
	assert.DeepEqual(t, "b2", s.B)
	assert.DeepEqual(t, "c1", s.C)
}

func TestBind_ZeroValueBind(t *testing.T) {
	var s struct {
		A int     `query:"a"`
		B float64 `query:"b"`
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?a=&b")

	err := bind.Bind(req.Req, nil, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, 0, s.A)
	assert.DeepEqual(t, float64(0), s.B)
}

func TestBind_DefaultValueBind(t *testing.T) {
	var s struct {
		A int      `default:"15"`
		B float64  `query:"b" default:"17"`
		C []int    `default:"15"`
		D []string `default:"qwe"`
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com")

	err := bind.Bind(req.Req, nil, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, 15, s.A)
	assert.DeepEqual(t, float64(17), s.B)
	assert.DeepEqual(t, 15, s.C[0])
	assert.DeepEqual(t, "qwe", s.D[0])

	var d struct {
		D [2]string `default:"qwe"`
	}

	err = bind.Bind(req.Req, nil, &d)
	if err == nil {
		t.Fatal("expected err")
	}
}

func TestBind_TypedefType(t *testing.T) {
	type Foo string
	type Bar *int
	type T struct {
		T1 string `query:"a"`
	}
	type TT T

	var s struct {
		A  Foo `query:"a"`
		B  Bar `query:"b"`
		T1 TT
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?a=1&b=2")
	err := bind.Bind(req.Req, nil, &s)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, Foo("1"), s.A)
	assert.DeepEqual(t, 2, *s.B)
	assert.DeepEqual(t, "1", s.T1.T1)
}

type CustomizedDecode struct {
	A string
}

func (c *CustomizedDecode) CustomizedFieldDecode(req *protocol.Request, params PathParams) error {
	q1 := req.URI().QueryArgs().Peek("a")
	if len(q1) == 0 {
		return fmt.Errorf("can be nil")
	}

	c.A = string(q1)
	return nil
}

func TestBind_CustomizedTypeDecode(t *testing.T) {
	type Foo struct {
		F ***CustomizedDecode
	}
	bind := Bind{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?a=1&b=2")
	result := Foo{}
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Fatal(err)
	}
	assert.DeepEqual(t, "1", (***result.F).A)

	type Bar struct {
		B *Foo
	}

	result2 := Bar{}
	err = bind.Bind(req.Req, nil, &result2)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, "1", (***(*result2.B).F).A)
}

func TestBind_JSON(t *testing.T) {
	bind := Bind{}
	type Req struct {
		J1 string `json:"j1"`
		J2 int    `json:"j2" query:"j2"` // 1. json unmarshal 2. query binding cover
		// todo: map
		J3 []byte    `json:"j3"`
		J4 [2]string `json:"j4"`
	}
	J3s := []byte("12")
	J4s := [2]string{"qwe", "asd"}

	req := newMockRequest().
		SetRequestURI("http://foobar.com?j2=13").
		SetJSONContentType().
		SetBody([]byte(fmt.Sprintf(`{"j1":"j1", "j2":12, "j3":[%d, %d], "j4":["%s", "%s"]}`, J3s[0], J3s[1], J4s[0], J4s[1])))
	var result Req
	err := bind.Bind(req.Req, nil, &result)
	if err != nil {
		t.Error(err)
	}
	assert.DeepEqual(t, "j1", result.J1)
	assert.DeepEqual(t, 13, result.J2)
	for idx, val := range J3s {
		assert.DeepEqual(t, val, result.J3[idx])
	}
	for idx, val := range J4s {
		assert.DeepEqual(t, val, result.J4[idx])
	}
}

func Benchmark_V2(b *testing.B) {
	bind := Bind{}
	type Req struct {
		Version string `path:"v"`
		ID      int    `query:"id"`
		Header  string `header:"h"`
		Form    string `form:"f"`
	}

	req := newMockRequest().
		SetRequestURI("http://foobar.com?id=12").
		SetHeaders("H", "header").
		SetPostArg("f", "form").
		SetUrlEncodeContentType()

	var params param.Params
	params = append(params, param.Param{
		Key:   "v",
		Value: "1",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Req
		err := bind.Bind(req.Req, params, &result)
		if err != nil {
			b.Error(err)
		}
		if result.ID != 12 {
			b.Error("Id failed")
		}
		if result.Form != "form" {
			b.Error("form failed")
		}
		if result.Header != "header" {
			b.Error("header failed")
		}
		if result.Version != "1" {
			b.Error("path failed")
		}
	}
}

func Benchmark_V1(b *testing.B) {
	type Req struct {
		Version string `path:"v"`
		ID      int    `query:"id"`
		Header  string `header:"h"`
		Form    string `form:"f"`
	}

	req := newMockRequest().
		SetRequestURI("http://foobar.com?id=12").
		SetHeaders("h", "header").
		SetPostArg("f", "form").
		SetUrlEncodeContentType()
	var params param.Params
	params = append(params, param.Param{
		Key:   "v",
		Value: "1",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Req
		err := binding.Bind(req.Req, &result, params)
		if err != nil {
			b.Error(err)
		}
		if result.ID != 12 {
			b.Error("Id failed")
		}
		if result.Form != "form" {
			b.Error("form failed")
		}
		if result.Header != "header" {
			b.Error("header failed")
		}
		if result.Version != "1" {
			b.Error("path failed")
		}
	}
}
