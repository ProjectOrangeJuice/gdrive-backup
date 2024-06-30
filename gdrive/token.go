package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"golang.org/x/oauth2"
)

func handleToken(token string, config *oauth2.Config) {
	exchangeToken, err := config.Exchange(context.TODO(), token)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	saveToken("token.json", exchangeToken)
}

func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
