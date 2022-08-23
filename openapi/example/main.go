package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"wago/openapi"

	"github.com/getkin/kin-openapi/openapi3"
)

// provide some defaults which are used across all the API
var (
	info = openapi3.Info{
		Version:     "1.0.0",
		Title:       "Test API",
		Description: "This is a test API.\n",
		Contact: &openapi3.Contact{
			Email: "john@doe.com",
		},
	}
	servers = openapi3.Servers{
		{
			Description: "Test server",
			URL:         "{schema}://{address}",
			Variables: map[string]*openapi3.ServerVariable{
				"address": {Default: "acme.com"},
				"schema":  {Default: "https", Enum: []string{"http", "https"}},
			},
		},
	}
	xAppIDParam = openapi.NewInHeaderParam("X-App-ID", "X-App-ID", "The application ID used to assign a request to.")
)

var _ openapi.Enum = (*TestEnum)(nil)

type TestEnum int

const (
	TestEnum_ValueOne TestEnum = iota
	TestEnum_ValueTwo
	TestEnum_ValueThree
)

func (e TestEnum) OpenApiValues() []interface{} {
	return []interface{}{
		"ValueOne",
		"ValueTwo",
		"ValueThree",
	}
}

type TestRequest struct {
	ReqProp1 string `json:"req_prop1" descr:"ReqProp1 must be set since the backend useses it to infer some important business logic."`
	ReqProp2 string `json:"req_prop2,omitempty"`
}

type EmbeddedEmbeddedModel struct {
	DeepProp1 string `json:"deep_prop1,omitempty"`
}

type EmbeddedModel struct {
	EmbeddedProp1 int                   `json:"embedded_prop1" descr:"this is a description of embedded property 1. This is a very important property."`
	EmbeddedProp2 string                `json:"embedded_prop2"`
	EmbeddedProp3 EmbeddedEmbeddedModel `json:"embedded_prop3"`
}

type TestResponse struct {
	RespProp1 string        `json:"resp_prop1"`
	RespProp2 float64       `json:"resp_prop2"`
	RespProp3 EmbeddedModel `json:"resp_prop3"`
	TestEnum  TestEnum      `json:"test_enum"`
}

type TestAPI struct{}

func (t *TestAPI) GetPaths() []*openapi.Path {
	tag1 := []string{"tag1"}
	tag2 := []string{"tag2"}
	tags := [][]string{tag1, tag2}

	return []*openapi.Path{
		{
			Template: "/api/v1/test",
			Operations: []*openapi.Operation{
				openapi.NewOperation(http.MethodPost, nil, tags[rand.Intn(len(tags))], "TestPost",
					openapi.NewRequestBody(openapi.SchemeType_Object, TestRequest{}),
					[]*openapi.Parameter{xAppIDParam},
					&openapi.WithBearerAuth,
					openapi.NewResponse(http.StatusOK, openapi.SchemeType_Object, TestResponse{}, nil,
						&openapi.ResponseMeta{
							Name:        "TestPostResponse",
							Description: "Returns the object after POSTing it to the server.",
						}),
				),
			},
		},
	}
}

func main() {
	ctx := context.Background()
	b := openapi.NewBuilder(ctx, "3.0.3", &openapi.BuilderConfig{
		API:     &TestAPI{},
		Info:    &info,
		Servers: &servers,
	})
	err := b.Build()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	bytes, err := b.Spec.ToBytes()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := os.WriteFile("./openapi/example/generated_example.json", bytes, 0644); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(0)
}