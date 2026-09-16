package main

import (
	"context"
	"log"

	remnawave "github.com/unser-elefant/remnawave-go"
)

func main() {
	client, err := remnawave.New(remnawave.Config{
		APIURL: "https://panel.example.com",
		APIKey: "your-api-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	user, err := client.Users().GetUserByTelegramID(context.Background(), 12345)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(user.Username)
}
