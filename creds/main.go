package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

func main() {
	getTokenFromWeb()
	handlers()
}

func handlers() {
	http.HandleFunc("/oauth2callback", handleOAuth2Callback)
	log.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil)) // Adjust the port if needed
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	// 1. Get Authorization Code from Query Parameters
	queryValues := r.URL.Query()
	authCode := queryValues.Get("code")

	if authCode == "" {
		fmt.Fprintf(w, "Error: Authorization code not found.")
		return
	}

	// 2. Print the Authorization Code (or Adapt for Token Exchange)
	fmt.Fprintf(w, "Received authorization code:\n%s\n\n", authCode)

	fmt.Printf("Received authorization code:\n%s\n\n", authCode)
	os.Exit(0)
}

// Request a token from the web
func getTokenFromWeb() {
	b, err := os.ReadFile("../creds.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
}
