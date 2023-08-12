package server

type Config struct {

	// Server Settings:
	IPv6             bool   `yaml:"ipv6" json:"ipv6" validate:"-"`
	Address          string `yaml:"address" json:"address" validate:"-"`
	Port             int    `yaml:"port" json:"port" validate:"required"`
	HTTPSPort        int    `yaml:"https_port" json:"https_port" validate:"-"`
	HTTPSPrivateKey  string `yaml:"https_key" json:"https_key" validate:"-"`
	HTTPSCertificate string `yaml:"https_cert" json:"https_cert" validate:"-"`

	// API:
	APIUrl          string `yaml:"openai_api_url" json:"openai_api_url" validate:"-"`
	APIKey          string `yaml:"openai_api_key" json:"openai_api_key" validate:"-"`
	APIOrganization string `yaml:"openai_api_org" json:"openai_api_org" validate:"-"`
	//
	//IntervalSeconds int    `yaml:"intervalSeconds" json:"intervalSeconds" bson:"intervalSeconds" validate:"required"`
	//Model           string `yaml:"model" json:"model" bson:"model" validate:"required"`
	//MaxLength       int    `yaml:"maxLength" json:"maxLength" bson:"maxLength" validate:"required"`
	//Cors            bool   `yaml:"cors" json:"cors" bson:"cors" validate:""`
}
