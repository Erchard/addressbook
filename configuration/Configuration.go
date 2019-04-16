package configuration

import (
	"../book"
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Seed []string
}

func Init() {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	timestamp := uint64(1555386491)

	seedstatus := book.NodeStatus{
		Address: &configuration.Seed[0],
		Status:  &timestamp,
	}
	book.Update(&seedstatus)
}
