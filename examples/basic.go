//go:build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/befship/fedex-sdk/fedex"
)

func main() {
	client := fedex.New(fedex.Config{
		ClientID:     os.Getenv("FEDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("FEDEX_CLIENT_SECRET"),
		Environment:  fedex.EnvSandbox,
	})
	tok, err := client.AccessToken(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("token ok, len=", len(tok), "base=", fedex.BaseURL(client.Environment()))
}
