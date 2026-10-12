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
		},
	}
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
	if o.openAPI.Paths == nil {
		o.openAPI.Paths = map[string]openapi.PathItem{}
	}

	inParameters, inRequestBody := o.parseInType(ep.inType)
	outResponses := o.parseOutType(ep.outType, ep.errorStatusCodes)

	operation := openapi.Operation{
		OperationID: newOperationID(ep.method, ep.path),
		Summary:     ep.summary,
		Description: ep.description,
		Tags:        ep.tags,
		Parameters:  inParameters,
		RequestBody: inRequestBody,
		Responses:   outResponses,
		Deprecated:  ep.deprecated,
	}

	pathItem := o.openAPI.Paths[ep.path] // other methods may already be registered under the same path

	switch ep.method {
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
		pathItem.AdditionalOperations[ep.method] = operation
	}

	o.openAPI.Paths[ep.path] = pathItem
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

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct && t.NumField() == 0 {
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

	for inType.Kind() == reflect.Pointer {
		return o.parseInType(inType.Elem())
	}

	if inType.Kind() != reflect.Struct {
		return nil, nil
	}

	var hasJSONBody, hasXMLBody, bodyRequired bool

	for field := range inType.Fields() {
		if !field.IsExported() {
			continue // skip private fields
		}

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

		if _, ok := bodyFieldName(field, "json"); ok {
			hasJSONBody = true
			bodyRequired = bodyRequired || isRequired(field)
		}

		if _, ok := bodyFieldName(field, "xml"); ok {
			hasXMLBody = true
			bodyRequired = bodyRequired || isRequired(field)
		}
	}

	if !hasJSONBody && !hasXMLBody {
		return parameters, nil
	}

	reqBody = &openapi.RequestBody{
		Content:  map[string]openapi.MediaType{},
		Required: bodyRequired,
	}

	if hasJSONBody {
		reqBody.Content["application/json"] = openapi.MediaType{
			Schema: o.structSchemaFor(inType, "json"),
		}
	}

	if hasXMLBody {
		reqBody.Content["application/xml"] = openapi.MediaType{
			Schema: o.structSchemaFor(inType, "xml"),
		}
	}

	return parameters, reqBody
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

func (o *oapi) structSchemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	name := o.schemaName(t, bodyTag)
	ref := openapi.SchemaSimpleOrRef{Ref: "#/components/schemas/" + name}

	if _, exists := o.openAPI.Components.Schemas[name]; exists {
		return ref
	}

	schema := openapi.Schema{
		Type:       openapi.DataTypeObject,
		Properties: map[string]openapi.SchemaSimpleOrRef{},
	}
	o.openAPI.Components.Schemas[name] = schema // placeholder against recursive types

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

	o.openAPI.Components.Schemas[name] = schema
	return ref
}

func (o *oapi) schemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	nullable := false
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		nullable = true
	}

	// user defined overrides are called on a pointer to zero value,
	// since its method set includes value receiver methods too
	zero := reflect.New(t).Interface()
	formatter, hasFormat := zero.(OpenAPIFormatter)

	var schema openapi.SchemaSimpleOrRef
	if typer, ok := zero.(OpenAPITyper); ok {
		schema.Type = openapi.DataType(typer.OpenAPIType())
	} else {
		schema = o.baseSchemaFor(t, bodyTag)
	}

	if hasFormat && schema.Ref == "" {
		schema.Format = openapi.DataFormat(formatter.OpenAPIFormat())
	}

	schema.Nullable = nullable
	return schema
}

// baseSchemaFor builds a schema for a non-pointer type, without user defined overrides.
func (o *oapi) baseSchemaFor(t reflect.Type, bodyTag string) openapi.SchemaSimpleOrRef {
	schema := openapi.SchemaSimpleOrRef{}

	switch t {
	case reflect.TypeFor[time.Time]():
		schema.Type = openapi.DataTypeString
		schema.Format = openapi.DataFormatDateTime
		return schema
	case reflect.TypeFor[[]byte]():
		schema.Type = openapi.DataTypeString
		schema.Format = openapi.DataFormatByte
		return schema
	}

	// popular third party types, recognized by name to avoid importing them
	switch t.PkgPath() + "." + t.Name() {
	case "github.com/google/uuid.UUID",
		"github.com/gofrs/uuid.UUID",
		"github.com/gofrs/uuid/v5.UUID",
		"github.com/satori/go.uuid.UUID":
		schema.Type = openapi.DataTypeString
		schema.Format = openapi.DataFormatUUID
		return schema
	case "cloud.google.com/go/civil.Date":
		schema.Type = openapi.DataTypeString
		schema.Format = openapi.DataFormatDate
		return schema
	case "github.com/shopspring/decimal.Decimal":
		schema.Type = openapi.DataTypeString
		schema.Format = openapi.DataFormatDecimal
		return schema
	}

	if reflect.PointerTo(t).Implements(reflect.TypeFor[encoding.TextMarshaler]()) {
		schema.Type = openapi.DataTypeString
		return schema
	}

	switch t.Kind() {
	case reflect.String:
		schema.Type = openapi.DataTypeString
	case reflect.Bool:
		schema.Type = openapi.DataTypeBoolean
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16:
		schema.Type = openapi.DataTypeInteger
		schema.Format = openapi.DataFormatInt32
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64:
		schema.Type = openapi.DataTypeInteger
		schema.Format = openapi.DataFormatInt64
	case reflect.Float32:
		schema.Type = openapi.DataTypeNumber
		schema.Format = openapi.DataFormatFloat
	case reflect.Float64:
		schema.Type = openapi.DataTypeNumber
		schema.Format = openapi.DataFormatDouble
	case reflect.Slice, reflect.Array:
		items := o.schemaFor(t.Elem(), bodyTag)
		schema.Type = openapi.DataTypeArray
		schema.Items = &items
	case reflect.Map:
		values := o.schemaFor(t.Elem(), bodyTag)
		schema.Type = openapi.DataTypeObject
		schema.AdditionalProperties = &values
	case reflect.Struct:
		return o.structSchemaFor(t, bodyTag)
	default:
		// interfaces, funcs, channels etc. can't be described, leave the schema empty (any type)
	}

	return schema
}

func (o *oapi) schemaName(t reflect.Type, bodyTag string) string {
	if o.openAPI.Components.Schemas == nil {
		o.openAPI.Components.Schemas = map[string]openapi.Schema{}
	}

	// package qualified name, e.g. civil.Date. Generic types (e.g. pkg.Page[github.com/x/y.User])
	// contain characters not allowed in component names, so they are replaced with `_`
	name := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, t.String())

	if t.Name() == "" {
		name = fmt.Sprintf("Anonymous%08x", rand.Uint32())
	}

	if bodyTag == "xml" {
		name += "XML"
	}

	return name
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

func newOperationID(method, path string) string {
	return fmt.Sprintf("%s %s", method, path)
}
