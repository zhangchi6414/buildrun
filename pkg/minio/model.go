package minio

type MinioOption struct {
	Endpoint        string `json:"endpoint,omitempty" `
	DisableSSL      bool   `json:"disableSSL,omitempty"`
	ForcePathStyle  string `json:"forcePathStyle,omitempty" `
	AccessKeyID     string `json:"accessKeyID,omitempty" `
	SecretAccessKey string `json:"secretAccessKey,omitempty" `
	SessionToken    string `json:"sessionToken,omitempty" `
	Bucket          string `json:"bucket,omitempty" `
	CodeName        string `json:"codeName,omitempty"`
}
