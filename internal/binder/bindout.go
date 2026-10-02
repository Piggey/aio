package binder

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
)

func (b *Binder) BindOut(w http.ResponseWriter, r *http.Request, obj any) error {
	accept := r.Header.Get("Accept")
	mimetype, _, err := mime.ParseMediaType(accept)
	if err != nil {
		return fmt.Errorf("mime.ParseMediaType: %w", err)
	}

	switch mimetype {
	case mimetypeApplicationXml:
		return xml.NewEncoder(w).Encode(obj)

	default:
		return json.NewEncoder(w).Encode(obj)
	}
}
