package openapi

type DataFormat string

const (
	DataFormatInt32    DataFormat = "int32"
	DataFormatInt64    DataFormat = "int64"
	DataFormatFloat    DataFormat = "float"
	DataFormatDouble   DataFormat = "double"
	DataFormatByte     DataFormat = "byte"   // base64 encoded characters
	DataFormatBinary   DataFormat = "binary" // any sequence of octets
	DataFormatDate     DataFormat = "date"
	DataFormatDateTime DataFormat = "date-time"
	DataFormatPassword DataFormat = "password"
)
