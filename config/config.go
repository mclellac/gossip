// Package config provides the gossip configuration file structure,
// initialization and reading.
package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/kelseyhightower/envconfig"
)

// Config is a gossip configuration struct.
type Config struct {
	Address string `hcl:"address" envconfig:"GOSSIP_ADDRESS"`
	BaseURL string `hcl:"base_url" envconfig:"GOSSIP_BASE_URL"`
	Title   string `hcl:"title" envconfig:"GOSSIP_TITLE"`

	JWT struct {
		Secret string `hcl:"secret" envconfig:"GOSSIP_JWT_SECRET"`
	} `hcl:"jwt"`

	FileStorage struct {
		Type string `hcl:"type" envconfig:"GOSSIP_FILE_STORAGE_TYPE"`

		Local struct {
			Dir string `hcl:"dir" envconfig:"GOSSIP_FILE_STORAGE_LOCAL_DIR"`
		} `hcl:"local"`

		GoogleCloudStorage struct {
			ServiceAccountFile string `hcl:"service_account_file" envconfig:"GOSSIP_FILE_STORAGE_GCS_SERVICE_ACCOUNT_FILE"`
			Bucket             string `hcl:"bucket" envconfig:"GOSSIP_FILE_STORAGE_GCS_BUCKET"`
		} `hcl:"google_cloud_storage"`

		AmazonS3 struct {
			AccessKey string `hcl:"access_key" envconfig:"GOSSIP_FILE_STORAGE_S3_ACCESS_KEY"`
			SecretKey string `hcl:"secret_key" envconfig:"GOSSIP_FILE_STORAGE_S3_SECRET_KEY"`
			Region    string `hcl:"region" envconfig:"GOSSIP_FILE_STORAGE_S3_REGION"`
			Bucket    string `hcl:"bucket" envconfig:"GOSSIP_FILE_STORAGE_S3_BUCKET"`
		} `hcl:"amazon_s3"`
	} `hcl:"file_storage"`

	Store struct {
		Type string `hcl:"type" envconfig:"GOSSIP_STORE_TYPE"`

		PostgreSQL struct {
			Address     string `hcl:"address" envconfig:"GOSSIP_STORE_POSTGRESQL_ADDRESS"`
			Username    string `hcl:"username" envconfig:"GOSSIP_STORE_POSTGRESQL_USERNAME"`
			Password    string `hcl:"password" envconfig:"GOSSIP_STORE_POSTGRESQL_PASSWORD"`
			Database    string `hcl:"database" envconfig:"GOSSIP_STORE_POSTGRESQL_DATABASE"`
			SSLMode     string `hcl:"sslmode" envconfig:"GOSSIP_STORE_POSTGRESQL_SSLMODE"`
			SSLRootCert string `hcl:"sslrootcert" envconfig:"GOSSIP_STORE_POSTGRESQL_SSLROOTCERT"`
		} `hcl:"postgresql"`

		MySQL struct {
			Address  string `hcl:"address" envconfig:"GOSSIP_STORE_MYSQL_ADDRESS"`
			Username string `hcl:"username" envconfig:"GOSSIP_STORE_MYSQL_USERNAME"`
			Password string `hcl:"password" envconfig:"GOSSIP_STORE_MYSQL_PASSWORD"`
			Database string `hcl:"database" envconfig:"GOSSIP_STORE_MYSQL_DATABASE"`
		} `hcl:"mysql"`
	} `hcl:"store"`

	OAuth struct {
		Google struct {
			ClientID string `hcl:"client_id" envconfig:"GOSSIP_OAUTH_GOOGLE_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"GOSSIP_OAUTH_GOOGLE_SECRET"`
		} `hcl:"google"`

		Facebook struct {
			ClientID string `hcl:"client_id" envconfig:"GOSSIP_OAUTH_FACEBOOK_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"GOSSIP_OAUTH_FACEBOOK_SECRET"`
		} `hcl:"facebook"`

		Github struct {
			ClientID string `hcl:"client_id" envconfig:"GOSSIP_OAUTH_GITHUB_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"GOSSIP_OAUTH_GITHUB_SECRET"`
		} `hcl:"github"`
	} `hcl:"oauth"`
}

// ReadFile reads a gossip config from file.
func ReadFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	cfg := &Config{}
	err = hcl.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal hcl: %v", err)
	}

	prepare(cfg)
	return cfg, nil
}

// ReadEnv reads a gossip config from environment variables.
func ReadEnv() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %v", err)
	}
	prepare(cfg)
	return cfg, nil
}

func prepare(cfg *Config) {
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
}

// Init generates an initial config string.
func Init() (string, error) {
	buf := new(bytes.Buffer)
	err := tpl.Execute(buf, map[string]interface{}{
		"jwt_secret": GenKeyHex(32),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenKeyHex generates a crypto-random key with byte length byteLen
// and hex-encodes it to a string.
func GenKeyHex(byteLen int) string {
	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

var tpl = template.Must(template.New("initial-config").Parse(strings.TrimSpace(`
address  = "127.0.0.1:8080"
base_url = "https://example.com/forum"
title    = "gossip"

jwt {
  secret = "{{.jwt_secret}}"
}

file_storage {
  type = "local"

  local {
    dir = "./gossip_data/public/"
  }

  google_cloud_storage {
    service_account_file = ""
    bucket               = ""
  }

  amazon_s3 {
    access_key = ""
    secret_key = ""
    region     = ""
    bucket     = ""
  }
}

store {
  type = "postgresql"

  postgresql {
    address  = "127.0.0.1:5432"
    username = ""
    password = ""
    database = ""
    sslmode  = "disable"
  }

  mysql {
    address  = "127.0.0.1:3306"
    username = ""
    password = ""
    database = ""
  }
}

oauth {
  google {
    client_id = ""
    secret    = ""
  }

  facebook {
    client_id = ""
    secret    = ""
  }

  github {
    client_id = ""
    secret    = ""
  }
}
`)))
