package aio

import "github.com/Piggey/aio/internal/openapi"

type openapiOpt func(oapi *oapi)

func WithOpenAPIDescription(desc string) openapiOpt {
	return func(oapi *oapi) {
		oapi.openAPI.Info.Description = desc
	}
}

func WithOpenAPIContact(name, url, email string) openapiOpt {
	return func(oapi *oapi) {
		contact := openapi.Contact{
			Name:  name,
			URL:   url,
			Email: email,
		}

		oapi.openAPI.Info.Contact = &contact
	}
}

func WithOpenAPILicense(name, url string) openapiOpt {
	return func(oapi *oapi) {
		licence := openapi.License{
			Name: name,
			URL:  url,
		}

		oapi.openAPI.Info.License = &licence
	}
}
