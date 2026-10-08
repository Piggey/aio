package aio

import "github.com/Piggey/aio/internal/openapi"

type oapiOpt func(oapi *oapi)

func WithOpenAPIDescription(desc string) oapiOpt {
	return func(oapi *oapi) {
		oapi.openAPI.Info.Description = desc
	}
}

func WithOpenAPIContact(name, url, email string) oapiOpt {
	return func(oapi *oapi) {
		contact := openapi.Contact{
			Name:  name,
			URL:   url,
			Email: email,
		}

		oapi.openAPI.Info.Contact = &contact
	}
}

func WithOpenAPILicense(name, url string) oapiOpt {
	return func(oapi *oapi) {
		licence := openapi.License{
			Name: name,
			URL:  url,
		}

		oapi.openAPI.Info.License = &licence
	}
}
