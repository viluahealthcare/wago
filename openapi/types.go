package openapi

import (
	"fmt"
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

type ContentType string

func (c ContentType) String() string {
	return string(c)
}

const (
	ContentType_ApplicationJson ContentType = "application/json"
)

// SchemeType is a custom type to represent the type of an OpenAPI-schema.
type SchemeType string

func (s SchemeType) String() string {
	return string(s)
}

const (
	SchemeType_Array   SchemeType = "array"
	SchemeType_Object  SchemeType = "object"
	SchemeType_String  SchemeType = "string"
	SchemeType_Int     SchemeType = "integer"
	SchemeType_Boolean SchemeType = "boolean"
	SchemeType_Number  SchemeType = "number"
	SchemeType_Nil     SchemeType = ""
)

// ParameterLocation is a custom type to represent the location of a parameter.
type ParameterLocation string

func (p ParameterLocation) String() string {
	return string(p)
}

const (
	ParameterLocation_InHeader ParameterLocation = "header"
	ParameterLocation_InQuery  ParameterLocation = "query"
	ParameterLocation_InPath   ParameterLocation = "path"
	ParamterLocation_InCookie  ParameterLocation = "cookie"
)

// Parameter describes a parameter of an operation,
// which can be a path, query, or header.
type Parameter struct {
	ID          string
	Name        string
	SchemeType  SchemeType
	In          ParameterLocation
	Description string
	Required    bool
}

// SecurityType is a custom type to represent the security scheme of an OpenAPI-document.
type SecurityType string

const (
	SecurityType_ApiKey SecurityType = "apiKey"
	SecurityType_Http   SecurityType = "http"
)

type SecurityScheme string

const (
	SecurityScheme_Basic  SecurityScheme = "basic"
	SecurityScheme_Bearer SecurityScheme = "bearer"
	SecurityScheme_NoAuth SecurityScheme = "noAuth"
)

// SecurityParam is an object to represent a security parameter in an OpenApi-document.
type SecurityParam struct {
	ID          string
	Name        string
	In          ParameterLocation
	Description string
	// SecurityType is the type of security scheme (e.g. apiKey, http, "oauth2").
	SecurityType SecurityType
	SecurityScheme
	BearerFormat string
}

// ResponseHeader is an object to represent a response header in an OpenApi-document.
type ResponseHeader struct {
	Name        string
	SchemaType  SchemeType
	Description string
}

// RequestBody is an object to represent the request body of an OpenApi-document.
type RequestBody struct {
	SchemaType SchemeType
	Model      reflect.Type
}

// NewRequestBody instantiates a new RequestBody.
func NewRequestBody(s SchemeType, model interface{}) *RequestBody {
	return &RequestBody{
		SchemaType: s,
		Model:      reflect.TypeOf(model),
	}
}

// Content is an object to represent the content of an OpenApi-document.
type Content struct {
	SchemaType SchemeType
	Model      reflect.Type
	Type       ContentType
}

// Response meta is an object to describe a response in an OpenApi-document with a name
// and description.
type ResponseMeta struct {
	Name        string
	Description string
}

// Response is an object to represent a response in an OpenApi-document.
type Response struct {
	Code    int
	Meta    *ResponseMeta
	Content *Content
	Headers []*ResponseHeader
}

// NewResponse instantiates a new Response.
func NewResponse(statusCode int, schemaType SchemeType, model interface{}, headers []*ResponseHeader, meta *ResponseMeta) *Response {
	c := &Response{
		Code:    statusCode,
		Headers: headers,
		Meta:    meta,
	}
	if model != nil {
		c.Content = &Content{
			SchemaType: schemaType,
			Model:      reflect.TypeOf(model),
			Type:       ContentType_ApplicationJson,
		}
	}
	return c
}

// Operation represents an OpenAPI operation.
type Operation struct {
	// Method is the HTTP method the operation supports.
	Method string
	// Handler is the name of the function that handles this operation.
	// This is currently more of a dummy object to indicate which handler is used.
	Handler interface{}
	// Tags is a slice of strings indicating which area of the generated client should include this path.
	Tags []string
	// OperationID is a unique string used to identify this operation.
	OperationID string
	// Summary is a summary of the operation.
	Summary string
	// Description is a verbose description of the operation.
	Description string
	// RequestBody defines the struct the handler expects in the request body.
	RequestBody *RequestBody
	// Params is a list of parameters this operation can take / expects.
	Parameters []*Parameter
	// SecurityParam defines the security scheme used for this operation.
	SecurityParam *SecurityParam
	// Responses is a variadic argument that represent the responses this operation can return.
	Responses []*Response
}

// NewOperation instantiates a new Operation struct.
func NewOperation(method string, handler interface{}, tags []string, operationID string, reqBody *RequestBody, params []*Parameter, sec *SecurityParam, responses ...*Response) *Operation {
	r := make([]*Response, 0)
	r = append(r, responses...)
	return &Operation{
		Method:        method,
		Handler:       handler,
		Tags:          tags,
		OperationID:   operationID,
		RequestBody:   reqBody,
		Parameters:    params,
		Responses:     r,
		SecurityParam: sec,
	}
}

// Path
type Path struct {
	// Specifies the field after the base path to be used as route segment.
	Template   string
	Operations []*Operation
}

// isValid returns an error if the path is not valid.
func (p *Path) isValid() error {
	if p.Template == "" {
		return fmt.Errorf("path template is empty")
	}
	return nil
}

// Impl represents an implementation of a REST-API.
type Impl interface {
	GetPaths() []*Path
}

// Enum is an interface which must be implemented by types
// that represent an enum in an OpenApi-Document.
type Enum interface {
	OpenApiValues() []interface{}
}

func (b *builder) customizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	// Enumeration Customizer
	if t.Implements(reflect.TypeOf((*Enum)(nil)).Elem()) {
		schema.Type = "string"

		m, _ := t.MethodByName("OpenApiValues")
		in := make([]reflect.Value, m.Type.NumIn())
		for i := 0; i < m.Type.NumIn(); i++ {
			in[i] = reflect.Zero(m.Type.In(i))
		}
		res := m.Func.Call(in)
		schema.Enum = res[0].Interface().([]interface{})

		if has := b.enumCache[t.Name()]; !has {
			b.enumCache[t.Name()] = true
		}
	}
	return nil
}

func chainCustomizer(customizers ...openapi3gen.SchemaCustomizerFn) openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		for _, c := range customizers {
			if err := c(name, t, tag, schema); err != nil {
				return err
			}
		}
		return nil
	}
}
