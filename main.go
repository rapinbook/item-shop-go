package main

import (
	"fmt"

	"github.com/rapinbook/item-shop-go/config"
)

func main() {
	config := config.ConfigGetting()
	fmt.Printf("Config struct: %v", config.Database.SSLMode)
}
