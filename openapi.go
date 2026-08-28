package aio

const openAPIVersion = "3.0.3"

type (
	openapi struct {
		Components   components            `json:"components,omitempty"`
		Paths        map[string]pathItem   `json:"paths"`
		ExternalDocs *externalDoc          `json:"externalDocs,omitempty"`
		Info         info                  `json:"info"`
		OpenAPI      string                `json:"openapi"`
		Servers      []server              `json:"servers,omitempty"`
		Security     []map[string][]string `json:"security,omitempty"`
		Tags         []tag                 `json:"tags,omitempty"`
	}

	info struct {
		Title          string   `json:"title"`
		Description    string   `json:"description,omitempty"`
		TermsOfService string   `json:"termsOfService,omitempty"`
		Contact        *contact `json:"contact,omitempty"`
		License        *license `json:"license,omitempty"`
		Version        string   `json:"version"`
	}

	contact struct {
		Name  string `json:"name,omitempty"`
		URL   string `json:"url,omitempty"`
		Email string `json:"email,omitempty"`
	}

	license struct {
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	}

	server struct {
		Variables   map[string]serverVariable `json:"variables,omitempty"`
		URL         string                    `json:"url"`
		Description string                    `json:"description,omitempty"`
	}

	serverVariable struct {
		Default     string   `json:"default"`
		Description string   `json:"description,omitempty"`
		Enum        []string `json:"enum"`
	}

	components struct {
		Schemas         map[string]schema         `json:"schemas,omitempty"`
		Responses       map[string]response       `json:"responses,omitempty"`
		Parameters      map[string]param          `json:"parameters,omitempty"`
		Examples        map[string]example        `json:"examples,omitempty"`
		RequestBodies   map[string]requestBody    `json:"requestBodies,omitempty"`
		Headers         map[string]param          `json:"headers,omitempty"`
		SecuritySchemes map[string]securityScheme `json:"securitySchemes,omitempty"`
		Links           map[string]link           `json:"links,omitempty"`
		Callbacks       map[string]callback       `json:"callbacks,omitempty"`
	}

	pathItem struct {
		Summary     string     `json:"summary,omitempty"`
		Description string     `json:"description,omitempty"`
		Get         *operation `json:"get,omitempty"`
		Put         *operation `json:"put,omitempty"`
		Post        *operation `json:"post,omitempty"`
		Delete      *operation `json:"delete,omitempty"`
		Options     *operation `json:"options,omitempty"`
		Head        *operation `json:"head,omitempty"`
		Patch       *operation `json:"patch,omitempty"`
		Trace       *operation `json:"trace,omitempty"`
		Servers     []server   `json:"servers,omitempty"`
		Parameters  []param    `json:"parameters,omitempty"`
	}

	operation struct {
		ExternalDocs *externalDoc          `json:"externalDocs,omitempty"`
		RequestBody  *requestBody          `json:"requestBody,omitempty"`
		Responses    map[string]response   `json:"responses"`
		Callbacks    map[string]callback   `json:"callbacks,omitempty"`
		Summary      string                `json:"summary,omitempty"`
		Description  string                `json:"description,omitempty"`
		OperationID  string                `json:"operationId,omitempty"`
		Tags         []string              `json:"tags,omitempty"`
		Parameters   []param               `json:"parameters,omitempty"`
		Security     []map[string][]string `json:"security,omitempty"`
		Servers      []server              `json:"servers,omitempty"`
		Deprecated   bool                  `json:"deprecated,omitempty"`
	}

	schema struct {
		Properties           map[string]schema `json:"properties,omitempty"`
		AdditionalProperties *schema           `json:"additionalProperties,omitempty"`
		Items                *schema           `json:"items,omitempty"`
		Description          string            `json:"description,omitempty"`
		Example              string            `json:"example,omitempty"`
		Format               DataFormat        `json:"format,omitempty"`
		Type                 DataType          `json:"type,omitempty"`
		Enum                 []any             `json:"enum,omitempty"`
		Required             []string          `json:"required,omitempty"`
		Minimum              float64           `json:"minimum,omitempty"`
		Maximum              float64           `json:"maximum,omitempty"`
		MinLength            int               `json:"minLength,omitempty"`
		MaxLength            int               `json:"maxLength,omitempty"`
		MinItems             int               `json:"minItems,omitempty"`
		MaxItems             int               `json:"maxItems,omitempty"`
		Nullable             bool              `json:"nullable,omitempty"`
	}

	reference struct {
		Ref string `json:"$ref,omitempty"`
	}

	response struct {
		Headers     map[string]param     `json:"headers,omitempty"`
		Content     map[string]mediaType `json:"content,omitempty"`
		Links       map[string]link      `json:"links,omitempty"`
		Description string               `json:"description"`
	}

	param struct {
		Example       any                  `json:"example,omitempty"`
		Schema        *schema              `json:"schema,omitempty"`
		Content       map[string]mediaType `json:"content,omitempty"`
		Examples      map[string]example   `json:"examples,omitempty"`
		Description   string               `json:"description,omitempty"`
		Style         string               `json:"style,omitempty"`
		In            ParameterLocation    `json:"in"`
		Name          string               `json:"name"`
		Deprecated    bool                 `json:"deprecated,omitempty"`
		Explode       bool                 `json:"explode,omitempty"`
		AllowReserved bool                 `json:"allowReserved,omitempty"`
		Required      bool                 `json:"required"`
	}

	mediaType struct {
		Example  any                 `json:"example,omitempty"`
		Examples map[string]example  `json:"examples,omitempty"`
		Encoding map[string]encoding `json:"encoding,omitempty"`
		Schema   schema              `json:"schema,omitempty"`
	}

	refMediaType struct {
		Example  any                 `json:"example,omitempty"`
		Examples map[string]example  `json:"examples,omitempty"`
		Encoding map[string]encoding `json:"encoding,omitempty"`
		Schema   reference           `json:"schema,omitempty"`
	}

	encoding struct {
		ContentType   string           `json:"contentType,omitempty"`
		Headers       map[string]param `json:"headers,omitempty"`
		Style         string           `json:"style,omitempty"`
		Explode       bool             `json:"explode,omitempty"`
		AllowReserved bool             `json:"allowReserved,omitempty"`
	}

	example struct {
		reference

		Summary       string `json:"summary,omitempty"`
		Description   string `json:"description,omitempty"`
		Value         any    `json:"value,omitempty"`
		ExternalValue string `json:"externalValue,omitempty"`
	}

	requestBody struct {
		Content     map[string]mediaType `json:"content"`
		Description string               `json:"description,omitempty"`
		Required    bool                 `json:"required,omitempty"`
	}

	refRequestBody struct {
		Content     map[string]refMediaType `json:"content"`
		Description string                  `json:"description,omitempty"`
		Required    bool                    `json:"required,omitempty"`
	}

	securityScheme struct {
		reference

		Type             string            `json:"type"`
		Description      string            `json:"description,omitempty"`
		Name             string            `json:"name,omitempty"`             // `apiKey`
		In               ParameterLocation `json:"in,omitempty"`               // `apiKey`
		Scheme           string            `json:"scheme,omitempty"`           // `http`
		BearerFormat     string            `json:"bearerFormat,omitempty"`     // `http`
		Flows            oAuth2Flows       `json:"flows,omitempty"`            // `oauth2`
		OpenIDConnectURL string            `json:"openIdConnectUrl,omitempty"` // `openIdConnect`
	}

	oAuth2Flows struct {
		Implicit          *oAuth2Flow `json:"implicit,omitempty"`
		Password          *oAuth2Flow `json:"password,omitempty"`
		ClientCredentials *oAuth2Flow `json:"clientCredentials,omitempty"`
		AuthorizationCode *oAuth2Flow `json:"authorizationCode,omitempty"`
	}

	oAuth2Flow struct {
		Scopes           map[string]string `json:"scopes,omitempty"`
		AuthorizationURL string            `json:"authorizationUrl,omitempty"`
		TokenURL         string            `json:"tokenUrl,omitempty"`
		RefreshURL       string            `json:"refreshUrl,omitempty"`
	}

	link struct {
		Server     server            `json:"server,omitempty"`
		Parameters map[string]string `json:"parameters,omitempty"`
		reference
		OperationRef string `json:"operationRef,omitempty"`
		OperationID  string `json:"operationId,omitempty"`
		RequestBody  string `json:"requestBody,omitempty"`
		Description  string `json:"description,omitempty"`
	}

	callback map[string]pathItem

	tag struct {
		ExternalDocs *externalDoc `json:"externalDocs,omitempty"`
		Name         string       `json:"name"`
		Description  string       `json:"description,omitempty"`
	}

	externalDoc struct {
		URL         string `json:"url"`
		Description string `json:"description,omitempty"`
	}
)
