package cucumber

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Request struct {
	Name        string `yaml:"name"`
	NodesNumber int    `yaml:"nodes_num"`
	DNSName     string `yaml:"dns_name,omitempty"`
}

// ParseRequestFromFile returns the creation request from a particular path
func ParseRequestFromFile(path string) (*Request, error) {
	var err error
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseRequestFromBytes(configFile)
}

// ParseRequestFromBytes returns a Request from a slice of bytes
func ParseRequestFromBytes(configFile []byte) (*Request, error) {
	r := &Request{}
	var err error

	err = yaml.Unmarshal(configFile, &r)
	if err != nil {
		return nil, err
	}
	return r, err
}
