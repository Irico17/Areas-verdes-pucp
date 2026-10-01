// Package contracts defines interfaces implemented across layers.
package contracts

import "encoding/json"

// IExifService defines operations for EXIF filtering and real MIME sniffing.
type IExifService interface {
	FiltrarExif(raw json.RawMessage) (any, error)
	MimeReal(b []byte) (mime string, ext string, ok bool)
}
