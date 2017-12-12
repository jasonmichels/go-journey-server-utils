package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/jasonmichels/journey-registry/journey"
	"gopkg.in/go-playground/validator.v9"
)

// PUBLIC The location of public directory to load public assets not served by Journey registry
const PUBLIC = "./public"

// INDEX The location of the index file to load
const INDEX = "./public/index.html"

// LoadJourneyConfig from json file
func LoadJourneyConfig(path string) (*journey.Journey, error) {
	var j journey.Journey

	abs, err := filepath.Abs(path)
	if err != nil {
		return &j, err
	}

	content, err := ioutil.ReadFile(abs)
	if err != nil {
		return &j, err
	}

	if err := json.Unmarshal(content, &j); err != nil {
		return &j, err
	}

	validate := validator.New()
	if err := validate.Struct(j); err != nil {
		return &j, err
	}

	log.Println("Successfully loaded journey.json configuration")
	return &j, nil
}

// Getenv Get environment variable or default to the optional string
func Getenv(key string, optional string) string {
	val, ok := os.LookupEnv(key)

	if !ok {
		return optional
	}
	return val
}
