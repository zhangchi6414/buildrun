package api

type RunImage struct {
	jobtype string
}

type DockerConfig struct {
	// Endpoint is the docker network endpoint or socket
	Endpoint string `json:"endpoint,omitempty"`

	// CertFile is the certificate file path for a TLS connection
	CertFile string `json:"certFile,omitempty"`

	// KeyFile is the key file path for a TLS connection
	KeyFile string `json:"keyFile,omitempty"`

	// CAFile is the certificate authority file path for a TLS connection
	CAFile string `json:"caFile,omitempty"`

	// UseTLS indicates if TLS must be used
	UseTLS bool `json:"useTLS,omitempty"`

	// TLSVerify indicates if TLS peer must be verified
	TLSVerify bool `json:"tlsVerify,omitempty"`
}
