package aio

import (
	"encoding"
	"encoding/json"
	"fmt"
	"html"
	"math/rand/v2"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/Piggey/aio/internal/binder"
	"github.com/Piggey/aio/internal/openapi"
)

type oapi struct {
	isEnabled  bool
	uiPath     string
	schemaPath string
	openAPI    openapi.OpenAPI
}

func defaultOAPI(name string, version string) oapi {
	return oapi{
		isEnabled:  true,
		uiPath:     "/openapi/ui",
		schemaPath: "/openapi/json",
		openAPI: openapi.OpenAPI{
			OpenAPI: openapi.OpenAPIVersion,
			Info: openapi.Info{
				Title:   name,
				Version: version,
			},
			Paths: map[string]openapi.PathItem{},
			Components: openapi.Components{
				Schemas: map[string]openapi.Schema{},
			},
		},
	}
}

func (o *oapi) addSecurityScheme(name string, scheme openapi.SecurityScheme, opts []securitySchemeOpt) {
	for _, opt := range opts {
		opt(&scheme)
	}

	if o.openAPI.Components.SecuritySchemes == nil {
		o.openAPI.Components.SecuritySchemes = map[string]openapi.SecurityScheme{}
	}

	o.openAPI.Components.SecuritySchemes[name] = scheme
}

// schemaHandler serves the OpenAPI document. It has to be created after all the endpoints are added.
func (o *oapi) schemaHandler() http.Handler {
	schema, err := json.Marshal(o.openAPI)
	if err != nil {
		panic(fmt.Sprintf("aio: json.Marshal openapi schema: %v", err))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(schema)
	})
}

// uiHandler serves the Scalar API reference UI pointing at schemaPath.
func (o *oapi) uiHandler() http.Handler {
	page := fmt.Sprintf(`<!doctype html>
<html>
<head>
	<title>%s | Scalar</title>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
	<div id="app"></div>
	<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	<script>Scalar.createApiReference('#app', { url: %q })</script>
</body>
</html>`, html.EscapeString(o.openAPI.Info.Title), o.schemaPath)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	})
}

func (o *oapi) addEndpoint(ep endpoint) {
	pathItem := o.openAPI.Paths[ep.path] // other methods may already be registered under the same path
	setOperation(&pathItem, ep.method, o.newOperation(ep))
	o.openAPI.Paths[ep.path] = pathItem
}

func (o *oapi) newOperation(ep endpoint) openapi.Operation {
	parameters, requestBody := o.parseInType(ep.inType)

	return openapi.Operation{
		OperationID: newOperationID(ep.method, ep.path),
		Summary:     ep.summary,
		Description: ep.description,
		Tags:        ep.tags,
		Parameters:  parameters,
		RequestBody: requestBody,
		Responses:   o.parseOutType(ep.outType, ep.errorStatusCodes),
		Deprecated:  ep.deprecated,
	}
}

func setOperation(pathItem *openapi.PathItem, method string, operation openapi.Operation) {
	switch method {
	case http.MethodGet:
		pathItem.Get = &operation
	case http.MethodPut:
		pathItem.Put = &operation
	case http.MethodPost:
		pathItem.Post = &operation
	case http.MethodDelete:
		pathItem.Delete = &operation
	case http.MethodOptions:
		pathItem.Options = &operation
	case http.MethodHead:
		pathItem.Head = &operation
	case http.MethodPatch:
		pathItem.Patch = &operation
	case http.MethodTrace:
		pathItem.Trace = &operation
	case "QUERY":
		pathItem.Query = &operation
	default:
		if pathItem.AdditionalOperations == nil {
			pathItem.AdditionalOperations = map[string]openapi.Operation{}
		}
		pathItem.AdditionalOperations[method] = operation
	}
}

func newOperationID(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}

func (o *oapi) parseOutType(outType reflect.Type, errorStatusCodes []int) map[int]openapi.Response {
	responses := map[int]openapi.Response{
		http.StatusOK: o.newResponse(http.StatusOK, outType),
	}

	problemDetailsType := reflect.TypeFor[ProblemDetails]()
	for _, statusCode := range errorStatusCodes {
		responses[statusCode] = o.newResponse(statusCode, problemDetailsType)
	}

	return responses
}

// newResponse describes a response body the way binder.BindOut writes it:
// json by default, xml if requested with the Accept header.
func (o *oapi) newResponse(statusCode int, t reflect.Type) openapi.Response {
	response := openapi.Response{
		Description: http.StatusText(statusCode),
	}

	t, _ = derefType(t)
	if isEmptyStruct(t) {
		return response // struct{} - nothing worth describing
	}

	response.Content = map[string]openapi.MediaType{
		"application/json": {Schema: o.schemaFor(t, "json")},
		"application/xml":  {Schema: o.schemaFor(t, "xml")},
	}

	return response
}

func (o *oapi) parseInType(inType reflect.Type) (parameters []openapi.Param, reqBody *openapi.RequestBody) {
	if inType == nil {
		return nil, nil
	}

	inType, _ = derefType(inType)
	if inType.Kind() != reflect.Struct {
		return nil, nil
	}

	return o.parseParams(inType), o.parseRequestBody(inType)
}

func (o *oapi) parseParams(t reflect.Type) []openapi.Param {
	var parameters []openapi.Param

	for field := range t.Fields() {
		if !field.IsExported() {
			continue // skip private fields
		}

		parameters = append(parameters, o.fieldParams(field)...)
	}

	return parameters
}

// fieldParams returns a parameter for every parameter tag of the field.
func (o *oapi) fieldParams(field reflect.StructField) []openapi.Param {
	var parameters []openapi.Param

	if name := field.Tag.Get(binder.TagHeader); name != "" {
		parameters = append(parameters, o.newParam(field, name, openapi.ParameterLocationHeader))
	}

	if name := field.Tag.Get(binder.TagQuery); name != "" {
		parameters = append(parameters, o.newParam(field, name, openapi.ParameterLocationQuery))
	}

	if name := field.Tag.Get(binder.TagPath); name != "" {
		parameters = append(parameters, o.newParam(field, name, openapi.ParameterLocationPath))
	}

	if name := field.Tag.Get(binder.TagCookie); name != "" {
		parameters = append(parameters, o.newParam(field, name, openapi.ParameterLocationCookie))
	}

	return parameters
}

func (o *oapi) newParam(field reflect.StructField, name string, location openapi.ParameterLocation) openapi.Param {
	schema := o.schemaFor(field.Type, "")

	return openapi.Param{
		Name:     name,
		In:       location,
		Schema:   &schema,
		Required: location == openapi.ParameterLocationPath || isRequired(field),
	}
}

func (o *oapi) parseRequestBody(t reflect.Type) *openapi.RequestBody {
	hasJSONBody, jsonRequired := hasBody(t, "json")
	hasXMLBody, xmlRequired := hasBody(t, "xml")

	if !hasJSONBody && !hasXMLBody {
		return nil
	}

	reqBody := &openapi.RequestBody{
		Content:  map[string]openapi.MediaType{},
		Required: jsonRequired || xmlRequired,
	}

	if hasJSONBody {
		reqBody.Content["application/json"] = openapi.MediaType{
			Schema: o.structSchemaFor(t, "json"),
		}
	}

	if hasXMLBody {
		reqBody.Content["application/xml"] = openapi.MediaType{
			Schema: o.structSchemaFor(t, "xml"),
		}
	}

	return reqBody
}

// hasBody reports whether any field of the struct is a part of the body encoded with bodyTag
// and whether any of those fields is required.
func hasBody(t reflect.Type, bodyTag string) (has bool, required bool) {
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		if _, ok := bodyFieldName(field, bodyTag); !ok {
			continue
		}

		has = true
		required = required || isRequired(field)
	}

	return has, required
}

// structSchemaFor registers the struct as a component schema and returns a reference to it.
func (o *oapi) structSchemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	name := schemaName(t, bodyTag)
	ref := openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/" + name}

	if _, exists := o.openAPI.Components.Schemas[name]; exists {
		return ref
	}

	o.openAPI.Components.Schemas[name] = openapi.Schema{} // placeholder against recursive types
	o.openAPI.Components.Schemas[name] = o.objectSchema(t, bodyTag)

	return ref
}

func (o *oapi) objectSchema(t reflect.Type, bodyTag string) openapi.Schema {
	schema := openapi.Schema{
		Type:       openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{},
	}

	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		propName, ok := bodyFieldName(field, bodyTag)
		if !ok {
			continue
		}

		schema.Properties[propName] = o.schemaFor(field.Type, bodyTag)
		if isRequired(field) {
			schema.Required = append(schema.Required, propName)
		}
	}

	return schema
}

func (o *oapi) schemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	t, nullable := derefType(t)

	// user defined overrides are called on a pointer to zero value,
	// since its method set includes value receiver methods too
	zero := reflect.New(t).Interface()

	var schema openapi.SchemaSimpleOrRef
	if typer, ok := zero.(OpenAPITyper); ok {
		schema.Type = openapi.DataType(typer.OpenAPIType())
	} else {
		schema = o.baseSchemaFor(t, bodyTag)
	}

	if formatter, ok := zero.(OpenAPIFormatter); ok && schema.Ref == "" {
		schema.Format = openapi.DataFormat(formatter.OpenAPIFormat())
	}

	schema.Nullable = nullable
	return schema
}

// baseSchemaFor builds a schema for a non-pointer type, without user defined overrides.
func (o *oapi) baseSchemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	if schema, ok := wellKnownSchema(t); ok {
		return schema
	}

	if schema, ok := primitiveSchema(t.Kind()); ok {
		return schema
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		items := o.schemaFor(t.Elem(), bodyTag)
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeArray, Items: &items}
	case reflect.Map:
		values := o.schemaFor(t.Elem(), bodyTag)
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeObject, AdditionalProperties: &values}
	case reflect.Struct:
		return o.structSchemaFor(t, bodyTag)
	default:
		// interfaces, funcs, channels etc. can't be described, leave the schema empty (any type)
		return openapi.SchemaSimpleOrRef{}
	}
}

// wellKnownSchema describes types with a custom text representation.
func wellKnownSchema(t reflect.Type) (openapi.SchemaSimpleOrRef, bool) {
	switch t {
	case reflect.TypeFor[time.Time]():
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatDateTime}, true
	case reflect.TypeFor[[]byte]():
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatByte}, true
	}

	// popular third party types, recognized by name to avoid importing them
	switch t.PkgPath() + "." + t.Name() {
	case "github.com/google/uuid.UUID",
		"github.com/gofrs/uuid.UUID",
		"github.com/gofrs/uuid/v5.UUID",
		"github.com/satori/go.uuid.UUID":
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatUUID}, true
	case "cloud.google.com/go/civil.Date":
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatDate}, true
	case "github.com/shopspring/decimal.Decimal":
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString, Format: openapi.DataFormatDecimal}, true
	}

	if reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextMarshaler]()) {
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString}, true
	}

	return openapi.SchemaSimpleOrRef{}, false
}

func primitiveSchema(kind reflect.Kind) (openapi.SchemaSimpleOrRef, bool) {
	switch kind {
	case reflect.String:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeString}, true
	case reflect.Bool:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeBoolean}, true
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt32}, true
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeInteger, Format: openapi.DataFormatInt64}, true
	case reflect.Float32:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeNumber, Format: openapi.DataFormatFloat}, true
	case reflect.Float64:
		return openapi.SchemaSimpleOrRef{Type: openapi.DataTypeNumber, Format: openapi.DataFormatDouble}, true
	default:
		return openapi.SchemaSimpleOrRef{}, false
	}
}

func schemaName(t reflect.Type, bodyTag string) string {
	name := sanitizeSchemaName(t.String())
	if t.Name() == "" {
		name = fmt.Sprintf("Anonymous%08x", rand.Uint32())
	}

	if bodyTag == "xml" {
		name += "XML"
	}

	return name
}

// sanitizeSchemaName replaces characters not allowed in component names with `_`.
// Package qualified names (e.g. civil.Date) stay untouched, generic types
// (e.g. pkg.Page[github.com/x/y.User]) are affected.
func sanitizeSchemaName(name string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
}

func bodyFieldName(field reflect.StructField, bodyTag string) (string, bool) {
	if bodyTag == "" {
		return field.Name, true
	}

	value := field.Tag.Get(bodyTag)
	if value == "-" {
		return "", false
	}

	if value == "" && hasParamTag(field) {
		return "", false // endpoint parameter, not a part of the body
	}

	name, _, _ := strings.Cut(value, ",")
	if name == "" {
		return field.Name, true
	}

	return name, true
}

func hasParamTag(field reflect.StructField) bool {
	return field.Tag.Get(binder.TagHeader) != "" ||
		field.Tag.Get(binder.TagQuery) != "" ||
		field.Tag.Get(binder.TagPath) != "" ||
		field.Tag.Get(binder.TagCookie) != ""
}

func isRequired(field reflect.StructField) bool {
	for rule := range strings.SplitSeq(field.Tag.Get("validate"), ",") {
		if rule == "required" {
			return true
		}
	}

	return false
}

// derefType strips pointers from the type. The second return value is true if there were any.
func derefType(t reflect.Type) (reflect.Type, bool) {
	nullable := false
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		nullable = true
	}

	return t, nullable
}

func isEmptyStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && t.NumField() == 0
}
