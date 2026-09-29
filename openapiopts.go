package aio

import "github.com/Piggey/aio/internal/openapi"

type openapiOpt func(oapi *openapi.OpenAPI)

func WithOpenAPIDescription(desc string) openapiOpt {
	return func(oapi *openapi.OpenAPI) {
		oapi.Info.Description = desc
	}
}

func WithOpenAPIContact(name, url, email string) openapiOpt {
	return func(oapi *openapi.OpenAPI) {
		contact := openapi.Contact{
			Name:  name,
			URL:   url,
			Email: email,
		}

		oapi.Info.Contact = &contact
	}
}

func WithOpenAPILicense(name, url string) openapiOpt {
	return func(oapi *openapi.OpenAPI) {
		licence := openapi.License{
			Name: name,
			URL:  url,
		}

		oapi.Info.License = &licence
	}
}
