package openapi

const OpenAPIVersion = "3.2.1"

type (
	OpenAPI struct {
		OpenAPI      string                `json:"openapi"`
		Info         Info                  `json:"info"`
		ExternalDocs *ExternalDoc          `json:"externalDocs,omitempty"`
		Tags         []Tag                 `json:"tags,omitempty"`
		Servers      []Server              `json:"servers,omitempty"`
		Security     []map[string][]string `json:"security,omitempty"`
		Paths        map[string]PathItem   `json:"paths"`
		Components   Components            `json:"components"`
	}

	Components struct {
		Schemas         map[string]Schema         `json:"schemas,omitempty"`
		SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	}

	SecurityScheme struct {
		Type              SecuritySchemeType `json:"type"`
		Description       string             `json:"description,omitempty"`
		Name              string             `json:"name,omitempty"`              // `apiKey`
		In                ParameterLocation  `json:"in,omitempty"`                // `apiKey`
		Scheme            string             `json:"scheme,omitempty"`            // `http`
		BearerFormat      string             `json:"bearerFormat,omitempty"`      // `http`
		Flows             *OAuth2Flows       `json:"flows,omitempty"`             // `oauth2`
		OpenIDConnectURL  string             `json:"openIdConnectUrl,omitempty"`  // `openIdConnect`
		OAuth2MetadataURL string             `json:"oauth2MetadataUrl,omitempty"` // `oauth2`
		Deprecated        bool               `json:"deprecated,omitempty"`
	}

	OAuth2Flows struct {
		Implicit            *OAuth2Flow `json:"implicit,omitempty"`
		Password            *OAuth2Flow `json:"password,omitempty"`
		ClientCredentials   *OAuth2Flow `json:"clientCredentials,omitempty"`
		AuthorizationCode   *OAuth2Flow `json:"authorizationCode,omitempty"`
		DeviceAuthorization *OAuth2Flow `json:"deviceAuthorization,omitempty"`
	}

	OAuth2Flow struct {
		AuthorizationURL       string            `json:"authorizationUrl,omitempty"`       // `implicit`, `authorizationCode`
		DeviceAuthorizationURL string            `json:"deviceAuthorizationUrl,omitempty"` // `deviceAuthorization`
		TokenURL               string            `json:"tokenUrl,omitempty"`               // `password`, `clientCredentials`, `authorizationCode`, `deviceAuthorization`
		RefreshURL             string            `json:"refreshUrl,omitempty"`
		Scopes                 map[string]string `json:"scopes"` // required, may be empty
	}

	PathItem struct {
		Summary              string               `json:"summary,omitempty"`
		Description          string               `json:"description,omitempty"`
		Parameters           []Param              `json:"parameters,omitempty"`
		Get                  *Operation           `json:"get,omitempty"`
		Put                  *Operation           `json:"put,omitempty"`
		Post                 *Operation           `json:"post,omitempty"`
		Delete               *Operation           `json:"delete,omitempty"`
		Options              *Operation           `json:"options,omitempty"`
		Head                 *Operation           `json:"head,omitempty"`
		Patch                *Operation           `json:"patch,omitempty"`
		Trace                *Operation           `json:"trace,omitempty"`
		Query                *Operation           `json:"query,omitempty"`
		AdditionalOperations map[string]Operation `json:"additionalOperations,omitempty"`
	}

	Operation struct {
		OperationID string           `json:"operationId,omitempty"`
		Summary     string           `json:"summary,omitempty"`
		Description string           `json:"description,omitempty"`
		Tags        []string         `json:"tags,omitempty"`
		Parameters  []Param          `json:"parameters,omitempty"`
		RequestBody *RequestBody     `json:"requestBody,omitempty"`
		Responses   map[int]Response `json:"responses,omitempty"`
		Deprecated  bool             `json:"deprecated,omitempty"`
	}

	Response struct {
		Description string               `json:"description,omitempty"`
		Headers     map[string]Param     `json:"headers,omitempty"`
		Content     map[string]MediaType `json:"content,omitempty"`
	}

	Param struct {
		Name          string             `json:"name"`
		In            ParameterLocation  `json:"in"`
		Schema        *SchemaSimpleOrRef `json:"schema,omitempty"`
		Deprecated    bool               `json:"deprecated,omitempty"`
		Explode       bool               `json:"explode,omitempty"`
		AllowReserved bool               `json:"allowReserved,omitempty"`
		Required      bool               `json:"required"`
	}

	RequestBody struct {
		Description string               `json:"description,omitempty"`
		Content     map[string]MediaType `json:"content"`
		Required    bool                 `json:"required,omitempty"`
	}

	MediaType struct {
		Example any               `json:"example,omitempty"`
		Schema  SchemaSimpleOrRef `json:"schema"`
	}

	Schema struct {
		Type        DataType                     `json:"type,omitempty"`
		Format      DataFormat                   `json:"format,omitempty"`
		Description string                       `json:"description,omitempty"`
		Required    []string                     `json:"required,omitempty"`
		Properties  map[string]SchemaSimpleOrRef `json:"properties,omitempty"`
		Example     string                       `json:"example,omitempty"`
	}

	SchemaSimpleOrRef struct {
		// object
		Ref string `json:"$ref,omitempty"`

		// simple type
		Type   DataType   `json:"type,omitempty"`
		Format DataFormat `json:"format,omitempty"`

		// array
		Items *SchemaSimpleOrRef `json:"items,omitempty"`

		// map
		AdditionalProperties *SchemaSimpleOrRef `json:"additionalProperties,omitempty"`

		// number
		Minimum float64 `json:"minimum,omitempty"`
		Maximum float64 `json:"maximum,omitempty"`

		// string
		MinLength int `json:"minLength,omitempty"`
		MaxLength int `json:"maxLength,omitempty"`

		// array
		MinItems int `json:"minItems,omitempty"`
		MaxItems int `json:"maxItems,omitempty"`

		Nullable bool `json:"nullable,omitempty"`
	}

	Server struct {
		Variables   map[string]ServerVariable `json:"variables,omitempty"`
		URL         string                    `json:"url"`
		Description string                    `json:"description,omitempty"`
	}

	ServerVariable struct {
		Default     string   `json:"default"`
		Description string   `json:"description,omitempty"`
		Enum        []string `json:"enum"`
	}

	Tag struct {
		ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
		Name         string       `json:"name"`
		Description  string       `json:"description,omitempty"`
	}

	ExternalDoc struct {
		URL         string `json:"url"`
		Description string `json:"description,omitempty"`
	}

	Info struct {
		Title          string   `json:"title"`
		Version        string   `json:"version"`
		Description    string   `json:"description,omitempty"`
		TermsOfService string   `json:"termsOfService,omitempty"`
		Contact        *Contact `json:"contact,omitempty"`
		License        *License `json:"license,omitempty"`
	}

	Contact struct {
		Name  string `json:"name,omitempty"`
		URL   string `json:"url,omitempty"`
		Email string `json:"email,omitempty"`
	}

	License struct {
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	}
)
