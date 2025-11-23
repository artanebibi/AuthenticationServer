package setup_google_refresh_token

import (
	"context"
	"fmt"
	"log"
	_ "net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to this link in your browser:\n\n%v\n\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func main() {
	if _, err := os.Stat("credentials.json"); os.IsNotExist(err) {
		log.Fatal("ERROR: credentials.json not found!\nPlease download Desktop app credentials from Google Cloud Console.")
	}

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse credentials: %v", err)
	}

	tokFile := "token.json"

	if _, err := os.Stat(tokFile); err == nil {
		fmt.Println("\ntoken.json already exists!")
		fmt.Print("Do you want to regenerate it? (y/n): ")
		var answer string
		fmt.Scan(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Setup cancelled.")
			return
		}
	}

	tok := getTokenFromWeb(config)
	saveToken(tokFile, tok)

	client := config.Client(context.Background(), tok)
	_, err = gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Failed to create Gmail service: %v", err)
	}

}
