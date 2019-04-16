package configuration

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	Seed          []string
	PreferredPort *string
	DbPath        string
}

var Config Configuration

func init() {
	file, err := os.Open("conf.json")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Config updated")
}
