package aio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Piggey/aio/internal/openapi"
)

func fieldOf[T any](t *testing.T, name string) reflect.StructField {
	t.Helper()
	field, ok := reflect.TypeFor[T]().FieldByName(name)
	if !ok {
		t.Fatalf("field %s not found", name)
	}
	return field
}

func newTestOAPI() *oapi {
	o := defaultOAPI("test", "v1")
	return &o
}

var (
	stringSchema = openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString}
	intSchema    = openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt64}
)

// test types

type textType struct{}

func (textType) MarshalText() ([]byte, error) { return nil, nil }

type emailType string

func (emailType) OpenAPIFormat() string { return "email" }

type moneyType struct{ amount int64 }

func (*moneyType) OpenAPIType() string   { return "string" }
func (*moneyType) OpenAPIFormat() string { return "decimal" }

type formattedStruct struct{}

func (formattedStruct) OpenAPIFormat() string { return "ignored" }

type item struct {
	Name string `json:"name" xml:"itemName"`
}

type page[T any] struct {
	Items []T `json:"items"`
}

type node struct {
	Next *node `json:"next"`
}

type params struct {
	ID      int      `path:"id"`
	Token   string   `header:"X-Token" validate:"required"`
	Limit   int      `query:"limit"`
	Tags    []string `query:"tag"`
	Session string   `cookie:"session"`
	private string   `query:"private"`
}

type body struct {
	Name    string `json:"name" xml:"name" validate:"required"`
	Ignored string `json:"-" xml:"-"`
}

type jsonOnlyBody struct {
	Name string `json:"name" xml:"-"`
}

type mixed struct {
	ID   int    `path:"id"`
	Name string `json:"name"`
}

// helpers

func TestDerefType(t *testing.T) {
	tests := []struct {
		name         string
		in           reflect.Type
		wantType     reflect.Type
		wantNullable bool
	}{
		{"value", reflect.TypeFor[int](), reflect.TypeFor[int](), false},
		{"pointer", reflect.TypeFor[*int](), reflect.TypeFor[int](), true},
		{"double pointer", reflect.TypeFor[**int](), reflect.TypeFor[int](), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotNullable := derefType(tt.in)
			assert.Equal(t, tt.wantType, gotType)
			assert.Equal(t, tt.wantNullable, gotNullable)
		})
	}
}

func TestIsEmptyStruct(t *testing.T) {
	assert.True(t, isEmptyStruct(reflect.TypeFor[struct{}]()))
	assert.False(t, isEmptyStruct(reflect.TypeFor[item]()))
	assert.False(t, isEmptyStruct(reflect.TypeFor[int]()))
}

func TestNewOperationID(t *testing.T) {
	assert.Equal(t, "GET /users/{id}", newOperationID(http.MethodGet, "/users/{id}"))
}

func TestIsRequired(t *testing.T) {
	type s struct {
		Required    string `validate:"required"`
		RequiredMax string `validate:"max=10,required"`
		Optional    string `validate:"max=10"`
		RequiredIf  string `validate:"required_if=Optional x"`
		NoTag       string
	}

	assert.True(t, isRequired(fieldOf[s](t, "Required")))
	assert.True(t, isRequired(fieldOf[s](t, "RequiredMax")))
	assert.False(t, isRequired(fieldOf[s](t, "Optional")))
	assert.False(t, isRequired(fieldOf[s](t, "RequiredIf")))
	assert.False(t, isRequired(fieldOf[s](t, "NoTag")))
}

func TestHasParamTag(t *testing.T) {
	assert.True(t, hasParamTag(fieldOf[params](t, "ID")))
	assert.True(t, hasParamTag(fieldOf[params](t, "Token")))
	assert.True(t, hasParamTag(fieldOf[params](t, "Limit")))
	assert.True(t, hasParamTag(fieldOf[params](t, "Session")))
	assert.False(t, hasParamTag(fieldOf[body](t, "Name")))
}

func TestBodyFieldName(t *testing.T) {
	type s struct {
		Tagged    string `json:"tagged"`
		OmitEmpty string `json:"omit,omitempty"`
		OnlyOpts  string `json:",omitempty"`
		Skipped   string `json:"-"`
		NoTag     string
		Param     string `query:"param"`
		ParamBody string `query:"pb" json:"paramBody"`
	}

	tests := []struct {
		field    string
		bodyTag  string
		wantName string
		wantOK   bool
	}{
		{"Tagged", "json", "tagged", true},
		{"OmitEmpty", "json", "omit", true},
		{"OnlyOpts", "json", "OnlyOpts", true},
		{"Skipped", "json", "", false},
		{"NoTag", "json", "NoTag", true},
		{"Param", "json", "", false},
		{"ParamBody", "json", "paramBody", true},
		{"Tagged", "xml", "Tagged", true},
		{"Skipped", "", "Skipped", true},
		{"Param", "", "Param", true},
	}

	for _, tt := range tests {
		t.Run(tt.field+"/"+tt.bodyTag, func(t *testing.T) {
			gotName, gotOK := bodyFieldName(fieldOf[s](t, tt.field), tt.bodyTag)
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestSanitizeSchemaName(t *testing.T) {
	assert.Equal(t, "civil.Date", sanitizeSchemaName("civil.Date"))
	assert.Equal(t, "my-type_1", sanitizeSchemaName("my-type_1"))
	assert.Equal(t, "pkg.Page_github.com_x_y.User_", sanitizeSchemaName("pkg.Page[github.com/x/y.User]"))
}

func TestSchemaName(t *testing.T) {
	assert.Equal(t, "aio.item", schemaName(reflect.TypeFor[item](), "json"))
	assert.Equal(t, "aio.itemXML", schemaName(reflect.TypeFor[item](), "xml"))
	assert.Equal(t, "aio.page_github.com_Piggey_aio.item_", schemaName(reflect.TypeFor[page[item]](), "json"))

	anonymous := schemaName(reflect.TypeFor[struct{ A int }](), "xml")
	assert.Regexp(t, `^Anonymous[0-9a-f]{8}XML$`, anonymous)
}

// schemas

func TestPrimitiveSchema(t *testing.T) {
	tests := []struct {
		kind   reflect.Kind
		want   openapi.SchemaSimpleOrRef
		wantOK bool
	}{
		{reflect.String, stringSchema, true},
		{reflect.Bool, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeBoolean}, true},
		{reflect.Int8, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt32}, true},
		{reflect.Int32, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt32}, true},
		{reflect.Uint16, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt32}, true},
		{reflect.Int, intSchema, true},
		{reflect.Uint32, intSchema, true},
		{reflect.Uint64, intSchema, true},
		{reflect.Float32, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeNumber, Format: openapi.DataFormatFloat}, true},
		{reflect.Float64, openapi.SchemaSimpleOrRef{Type: openapi.DataTypeNumber, Format: openapi.DataFormatDouble}, true},
		{reflect.Struct, openapi.SchemaSimpleOrRef{}, false},
		{reflect.Slice, openapi.SchemaSimpleOrRef{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			got, gotOK := primitiveSchema(tt.kind)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestWellKnownSchema(t *testing.T) {
	tests := []struct {
		name   string
		in     reflect.Type
		want   openapi.SchemaSimpleOrRef
		wantOK bool
	}{
		{"time", reflect.TypeFor[time.Time](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatDateTime}, true},
		{"bytes", reflect.TypeFor[[]byte](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatByte}, true},
		{"text marshaler", reflect.TypeFor[textType](), stringSchema, true},
		{"string", reflect.TypeFor[string](), openapi.SchemaSimpleOrRef{}, false},
		{"struct", reflect.TypeFor[item](), openapi.SchemaSimpleOrRef{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := wellKnownSchema(tt.in)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}

func TestSchemaFor(t *testing.T) {
	tests := []struct {
		name string
		in   reflect.Type
		want openapi.SchemaSimpleOrRef
	}{
		{"primitive", reflect.TypeFor[string](), stringSchema},
		{"pointer", reflect.TypeFor[*string](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Nullable: true}},
		{"slice", reflect.TypeFor[[]string](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeArray, Items: new(stringSchema)}},
		{"array", reflect.TypeFor[[2]string](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeArray, Items: new(stringSchema)}},
		{"map", reflect.TypeFor[map[string]string](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeObject, AdditionalProperties: new(stringSchema)}},
		{"any", reflect.TypeFor[any](), openapi.SchemaSimpleOrRef{}},
		{"struct", reflect.TypeFor[item](), openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.item"}},
		{"struct pointer", reflect.TypeFor[*item](), openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.item", Nullable: true}},
		{"formatter", reflect.TypeFor[emailType](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: "email"}},
		{"typer and formatter", reflect.TypeFor[*moneyType](), openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: "decimal", Nullable: true}},
		{"formatter on struct ref", reflect.TypeFor[formattedStruct](), openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.formattedStruct"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newTestOAPI().schemaFor(tt.in, "json"))
		})
	}
}

func TestSchemaFor_TyperSkipsComponent(t *testing.T) {
	o := newTestOAPI()
	o.schemaFor(reflect.TypeFor[moneyType](), "json")
	assert.Empty(t, o.openAPI.Components.Schemas)
}

type colorEnum string

type priorityEnum int

type withEnums struct {
	Color    colorEnum     `query:"color"`
	Priority *priorityEnum `json:"priority"`
	Colors   []colorEnum   `json:"colors"`
}

func TestWithEnumType(t *testing.T) {
	o := newTestOAPI()
	WithEnumType([]colorEnum{"red", "green"})(o)
	WithEnumType([]priorityEnum{1, 2, 3})(o)

	parameters, reqBody := o.parseInType(reflect.TypeFor[withEnums]())

	colorRef := openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.colorEnum"}
	assert.Equal(t, []openapi.Param{{Name: "color", In: openapi.ParameterLocationQuery, Schema: &colorRef}}, parameters)

	assert.Equal(t, openapi.Schema{
		Type: openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{
			"priority": {Ref: "#/components/schemas/aio.priorityEnum", Nullable: true},
			"colors":   {Type: openapi.DataTypeArray, Items: &colorRef},
		},
	}, o.openAPI.Components.Schemas["aio.withEnums"])
	assert.NotNil(t, reqBody)

	assert.Equal(t, openapi.Schema{
		Type: openapi.DataTypeString,
		Enum: []any{colorEnum("red"), colorEnum("green")},
	}, o.openAPI.Components.Schemas["aio.colorEnum"])

	assert.Equal(t, openapi.Schema{
		Type:   openapi.DataTypeInteger,
		Format: openapi.DataFormatInt64,
		Enum:   []any{priorityEnum(1), priorityEnum(2), priorityEnum(3)},
	}, o.openAPI.Components.Schemas["aio.priorityEnum"])
}

func TestWithEnumType_JSON(t *testing.T) {
	o := newTestOAPI()
	WithEnumType([]colorEnum{"red", "green"})(o)
	o.schemaFor(reflect.TypeFor[colorEnum](), "json")

	got, err := json.Marshal(o.openAPI.Components.Schemas["aio.colorEnum"])
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"string","enum":["red","green"]}`, string(got))
}

func TestStructSchemaFor(t *testing.T) {
	o := newTestOAPI()

	jsonRef := o.structSchemaFor(reflect.TypeFor[body](), "json")
	xmlRef := o.structSchemaFor(reflect.TypeFor[item](), "xml")

	assert.Equal(t, openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.body"}, jsonRef)
	assert.Equal(t, openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.itemXML"}, xmlRef)

	assert.Equal(t, openapi.Schema{
		Type:       openapi.DataTypeObject,
		Required:   []string{"name"},
		Properties: map[string]openapi.SchemaSimpleOrRef{"name": stringSchema},
	}, o.openAPI.Components.Schemas["aio.body"])

	assert.Equal(t, openapi.Schema{
		Type:       openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{"itemName": stringSchema},
	}, o.openAPI.Components.Schemas["aio.itemXML"])
}

func TestStructSchemaFor_Recursive(t *testing.T) {
	o := newTestOAPI()
	o.structSchemaFor(reflect.TypeFor[node](), "json")

	assert.Equal(t, openapi.Schema{
		Type: openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{
			"next": {Ref: "#/components/schemas/aio.node", Nullable: true},
		},
	}, o.openAPI.Components.Schemas["aio.node"])
}

func TestStructSchemaFor_ExistingIsNotOverwritten(t *testing.T) {
	o := newTestOAPI()
	existing := openapi.Schema{Description: "existing"}
	o.openAPI.Components.Schemas["aio.item"] = existing

	o.structSchemaFor(reflect.TypeFor[item](), "json")
	assert.Equal(t, existing, o.openAPI.Components.Schemas["aio.item"])
}

func TestObjectSchema_SkipsParams(t *testing.T) {
	got := newTestOAPI().objectSchema(reflect.TypeFor[mixed](), "json")

	assert.Equal(t, openapi.Schema{
		Type:       openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{"name": stringSchema},
	}, got)
}

// in type

func TestFieldParams(t *testing.T) {
	type multi struct {
		ID int `path:"id" query:"id"`
	}

	got := newTestOAPI().fieldParams(fieldOf[multi](t, "ID"))

	assert.Equal(t, []openapi.Param{
		{Name: "id", In: openapi.ParameterLocationQuery, Schema: new(intSchema)},
		{Name: "id", In: openapi.ParameterLocationPath, Schema: new(intSchema), Required: true},
	}, got)
}

func TestParseParams(t *testing.T) {
	got := newTestOAPI().parseParams(reflect.TypeFor[params]())

	assert.Equal(t, []openapi.Param{
		{Name: "id", In: openapi.ParameterLocationPath, Schema: new(intSchema), Required: true},
		{Name: "X-Token", In: openapi.ParameterLocationHeader, Schema: new(stringSchema), Required: true},
		{Name: "limit", In: openapi.ParameterLocationQuery, Schema: new(intSchema)},
		{Name: "tag", In: openapi.ParameterLocationQuery, Schema: new(openapi.SchemaSimpleOrRef{Type: openapi.DataTypeArray, Items: new(stringSchema)})},
		{Name: "session", In: openapi.ParameterLocationCookie, Schema: new(stringSchema)},
	}, got)
}

func TestHasBody(t *testing.T) {
	tests := []struct {
		name         string
		in           reflect.Type
		bodyTag      string
		wantHas      bool
		wantRequired bool
	}{
		{"required json", reflect.TypeFor[body](), "json", true, true},
		{"required xml", reflect.TypeFor[body](), "xml", true, true},
		{"skipped xml", reflect.TypeFor[jsonOnlyBody](), "xml", false, false},
		{"optional json", reflect.TypeFor[jsonOnlyBody](), "json", true, false},
		{"params only", reflect.TypeFor[params](), "json", false, false},
		{"empty", reflect.TypeFor[struct{}](), "json", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHas, gotRequired := hasBody(tt.in, tt.bodyTag)
			assert.Equal(t, tt.wantHas, gotHas)
			assert.Equal(t, tt.wantRequired, gotRequired)
		})
	}
}

func TestParseRequestBody(t *testing.T) {
	t.Run("no body", func(t *testing.T) {
		assert.Nil(t, newTestOAPI().parseRequestBody(reflect.TypeFor[params]()))
	})

	t.Run("json and xml", func(t *testing.T) {
		assert.Equal(t, &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.body"}},
				"application/xml":  {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.bodyXML"}},
			},
		}, newTestOAPI().parseRequestBody(reflect.TypeFor[body]()))
	})

	t.Run("json only", func(t *testing.T) {
		assert.Equal(t, &openapi.RequestBody{
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.jsonOnlyBody"}},
			},
		}, newTestOAPI().parseRequestBody(reflect.TypeFor[jsonOnlyBody]()))
	})
}

func TestParseInType(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		params, reqBody := newTestOAPI().parseInType(nil)
		assert.Nil(t, params)
		assert.Nil(t, reqBody)
	})

	t.Run("not a struct", func(t *testing.T) {
		params, reqBody := newTestOAPI().parseInType(reflect.TypeFor[int]())
		assert.Nil(t, params)
		assert.Nil(t, reqBody)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		params, reqBody := newTestOAPI().parseInType(reflect.TypeFor[*mixed]())
		if assert.Len(t, params, 1) {
			assert.Equal(t, "id", params[0].Name)
		}
		if assert.NotNil(t, reqBody) {
			assert.Equal(t, "#/components/schemas/aio.mixed", reqBody.Content["application/json"].Schema.Ref)
		}
	})
}

// out type

func TestNewResponse(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		got := newTestOAPI().newResponse(http.StatusOK, reflect.TypeFor[struct{}]())
		assert.Equal(t, openapi.Response{Description: "OK"}, got)
	})

	t.Run("struct", func(t *testing.T) {
		got := newTestOAPI().newResponse(http.StatusCreated, reflect.TypeFor[*item]())
		assert.Equal(t, openapi.Response{
			Description: "Created",
			Content: map[string]openapi.MediaType{
				"application/json": {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.item"}},
				"application/xml":  {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.itemXML"}},
			},
		}, got)
	})
}

func TestParseOutType(t *testing.T) {
	got := newTestOAPI().parseOutType(reflect.TypeFor[item](), []int{http.StatusBadRequest, http.StatusNotFound})

	problemDetails := map[string]openapi.MediaType{
		"application/json": {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.ProblemDetails"}},
		"application/xml":  {Schema: openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/aio.ProblemDetailsXML"}},
	}

	assert.Len(t, got, 3)
	assert.Equal(t, "#/components/schemas/aio.item", got[http.StatusOK].Content["application/json"].Schema.Ref)
	assert.Equal(t, openapi.Response{Description: "Bad Request", Content: problemDetails}, got[http.StatusBadRequest])
	assert.Equal(t, openapi.Response{Description: "Not Found", Content: problemDetails}, got[http.StatusNotFound])
}

// endpoints

func TestSetOperation(t *testing.T) {
	op := openapi.Operation{OperationID: "op"}

	tests := []struct {
		method string
		get    func(openapi.PathItem) *openapi.Operation
	}{
		{http.MethodGet, func(p openapi.PathItem) *openapi.Operation { return p.Get }},
		{http.MethodPut, func(p openapi.PathItem) *openapi.Operation { return p.Put }},
		{http.MethodPost, func(p openapi.PathItem) *openapi.Operation { return p.Post }},
		{http.MethodDelete, func(p openapi.PathItem) *openapi.Operation { return p.Delete }},
		{http.MethodOptions, func(p openapi.PathItem) *openapi.Operation { return p.Options }},
		{http.MethodHead, func(p openapi.PathItem) *openapi.Operation { return p.Head }},
		{http.MethodPatch, func(p openapi.PathItem) *openapi.Operation { return p.Patch }},
		{http.MethodTrace, func(p openapi.PathItem) *openapi.Operation { return p.Trace }},
		{"QUERY", func(p openapi.PathItem) *openapi.Operation { return p.Query }},
		{"LINK", func(p openapi.PathItem) *openapi.Operation {
			if op, ok := p.AdditionalOperations["LINK"]; ok {
				return &op
			}
			return nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			var pathItem openapi.PathItem
			setOperation(&pathItem, tt.method, op)
			assert.Equal(t, &op, tt.get(pathItem))
		})
	}
}

func TestNewOperation(t *testing.T) {
	got := newTestOAPI().newOperation(endpoint{
		method:           http.MethodPost,
		path:             "/items/{id}",
		summary:          "summary",
		description:      "description",
		tags:             []string{"items"},
		inType:           reflect.TypeFor[mixed](),
		outType:          reflect.TypeFor[struct{}](),
		errorStatusCodes: []int{http.StatusNotFound},
		deprecated:       true,
	})

	assert.Equal(t, "POST /items/{id}", got.OperationID)
	assert.Equal(t, "summary", got.Summary)
	assert.Equal(t, "description", got.Description)
	assert.Equal(t, []string{"items"}, got.Tags)
	assert.True(t, got.Deprecated)
	assert.Len(t, got.Parameters, 1)
	assert.NotNil(t, got.RequestBody)
	assert.Len(t, got.Responses, 2)
}

func TestAddEndpoint_SamePathDifferentMethods(t *testing.T) {
	o := newTestOAPI()
	empty := reflect.TypeFor[struct{}]()

	o.addEndpoint(endpoint{method: http.MethodGet, path: "/items", inType: empty, outType: empty})
	o.addEndpoint(endpoint{method: http.MethodPost, path: "/items", inType: empty, outType: empty})

	pathItem := o.openAPI.Paths["/items"]
	if assert.NotNil(t, pathItem.Get) && assert.NotNil(t, pathItem.Post) {
		assert.Equal(t, "GET /items", pathItem.Get.OperationID)
		assert.Equal(t, "POST /items", pathItem.Post.OperationID)
	}
}

// handlers

func TestSchemaHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestOAPI().schemaHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi/json", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got openapi.OpenAPI
	if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got)) {
		assert.Equal(t, "test", got.Info.Title)
	}
}

func TestUIHandler(t *testing.T) {
	o := newTestOAPI()
	o.openAPI.Info.Title = "<script>"

	rec := httptest.NewRecorder()
	o.uiHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi/ui", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `url: "/openapi/json"`)
	assert.Contains(t, rec.Body.String(), "<title>&lt;script&gt; | Scalar</title>")
}

// security schemes

func TestSecuritySchemes(t *testing.T) {
	o := newTestOAPI()
	scopes := map[string]string{"read": "read access"}

	for _, opt := range []oapiOpt{
		WithAPIKeySecurity("header", "X-API-Key", ParameterLocationHeader),
		WithAPIKeySecurity("query", "api_key", ParameterLocationQuery),
		WithAPIKeySecurity("cookie", "session", ParameterLocationCookie, WithSecuritySchemeDescription("desc"), WithSecuritySchemeDeprecated()),
		WithHTTPSecurity("digest", "digest"),
		WithBasicSecurity("basic"),
		WithBearerSecurity("bearer", "JWT"),
		WithMutualTLSSecurity("mtls"),
		WithOAuth2Security("oauth2",
			WithOAuth2MetadataURL("https://auth/.well-known/oauth-authorization-server"),
			WithOAuth2ImplicitFlow("https://auth/authorize", scopes),
			WithOAuth2PasswordFlow("https://auth/token", nil),
			WithOAuth2ClientCredentialsFlow("https://auth/token", scopes),
			WithOAuth2AuthorizationCodeFlow("https://auth/authorize", "https://auth/token", scopes, WithOAuth2RefreshURL("https://auth/refresh")),
			WithOAuth2DeviceAuthorizationFlow("https://auth/device", "https://auth/token", scopes),
		),
		WithOpenIDConnectSecurity("oidc", "https://auth/.well-known/openid-configuration"),
		WithBasicSecurity("ignoredFlow", WithOAuth2PasswordFlow("https://auth/token", nil)),
	} {
		opt(o)
	}

	assert.Equal(t, map[string]openapi.SecurityScheme{
		"header": {Type: openapi.SecuritySchemeTypeAPIKey, Name: "X-API-Key", In: openapi.ParameterLocationHeader},
		"query":  {Type: openapi.SecuritySchemeTypeAPIKey, Name: "api_key", In: openapi.ParameterLocationQuery},
		"cookie": {Type: openapi.SecuritySchemeTypeAPIKey, Name: "session", In: openapi.ParameterLocationCookie, Description: "desc", Deprecated: true},
		"digest": {Type: openapi.SecuritySchemeTypeHTTP, Scheme: "digest"},
		"basic":  {Type: openapi.SecuritySchemeTypeHTTP, Scheme: "basic"},
		"bearer": {Type: openapi.SecuritySchemeTypeHTTP, Scheme: "bearer", BearerFormat: "JWT"},
		"mtls":   {Type: openapi.SecuritySchemeTypeMutualTLS},
		"oauth2": {
			Type:              openapi.SecuritySchemeTypeOAuth2,
			OAuth2MetadataURL: "https://auth/.well-known/oauth-authorization-server",
			Flows: &openapi.OAuth2Flows{
				Implicit:            &openapi.OAuth2Flow{AuthorizationURL: "https://auth/authorize", Scopes: scopes},
				Password:            &openapi.OAuth2Flow{TokenURL: "https://auth/token", Scopes: map[string]string{}},
				ClientCredentials:   &openapi.OAuth2Flow{TokenURL: "https://auth/token", Scopes: scopes},
				AuthorizationCode:   &openapi.OAuth2Flow{AuthorizationURL: "https://auth/authorize", TokenURL: "https://auth/token", RefreshURL: "https://auth/refresh", Scopes: scopes},
				DeviceAuthorization: &openapi.OAuth2Flow{DeviceAuthorizationURL: "https://auth/device", TokenURL: "https://auth/token", Scopes: scopes},
			},
		},
		"oidc":        {Type: openapi.SecuritySchemeTypeOpenIDConnect, OpenIDConnectURL: "https://auth/.well-known/openid-configuration"},
		"ignoredFlow": {Type: openapi.SecuritySchemeTypeHTTP, Scheme: "basic"},
	}, o.openAPI.Components.SecuritySchemes)
}

func TestSecuritySchemes_JSON(t *testing.T) {
	o := newTestOAPI()
	WithOAuth2Security("oauth2", WithOAuth2ClientCredentialsFlow("https://auth/token", nil))(o)
	WithBasicSecurity("basic")(o)

	got, err := json.Marshal(o.openAPI.Components.SecuritySchemes)
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{
			"oauth2": {"type": "oauth2", "flows": {"clientCredentials": {"tokenUrl": "https://auth/token", "scopes": {}}}},
			"basic": {"type": "http", "scheme": "basic"}
		}`, string(got))
	}
}
