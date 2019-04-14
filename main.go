package main

import (
	"./book"
	"./configuration"
)

func main() {
	book.Start()
	configuration.Init()
}
