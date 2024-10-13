package core

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

func ReadConfig(path string) Config {
	contents, err := os.ReadFile(path)
	if err != nil {
		println("Unable to load config.toml")
		panic(err)
	}
	var config Config
	toml.Unmarshal([]byte(contents), &config)
	return config
}

type Config struct {
	// Integrate into a web server that hosts the deployed static files
	FastCgi *FastCgiConfig

	// Host your own web server that serves the deployed static files
	Http *HttpConfig

	// Deploy static files to a folder on the local filesystem. Used to deploy to the server that
	// fastcgi connects to.
	DeployToFolder *DeployToFolderConfig

	// Deploy static files to a folder on a remote filesystem using ssh. Used to deploy to the
	// server that fastcgi connects to.
	DeployToSSH *DeployToSSHConfig
}

type FastCgiConfig struct {
	Address string
}

type HttpConfig struct {
	Address string
}

type DeployToFolderConfig struct {
	Path string
}

type DeployToSSHConfig struct {
	Host    string
	User    string
	KeyFile string
}
