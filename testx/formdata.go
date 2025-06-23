package testx

// FileToUpload represents a file that will be part of the multipart form data.
type FileToUpload struct {
	FieldName string
	FileName  string
	Content   []byte
}

// FormData represents the data for a multipart/form-data request.
type FormData struct {
	Fields map[string]string
	Files  []FileToUpload
}
