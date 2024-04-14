package main

import (
	"fmt"

	"github.com/rapinbook/item-shop-go/config"
	"github.com/rapinbook/item-shop-go/databases"
	"github.com/rapinbook/item-shop-go/server"
)

func main() {
	// Echo instance
	config := config.ConfigGetting()
	fmt.Printf("Config struct: %v", config.Database.SSLMode)
	db := databases.NewPostgresDatabase(config.Database)
	s := server.NewEchoServer(db, config)
	s.Start()
}
