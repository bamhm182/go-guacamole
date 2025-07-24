package main

import (
 "context"
 "encoding/json"
 "fmt"
 "log"
 "net/http"
 "net/url"

 guacamole "github.com/bamhm182/go-guacamole/guacamole"
)

func main() {
 //Guacamole API base URL
 baseURL := "http://localhost:8080/api"

 // Parse the base URL
 parsedURL, err := url.Parse(baseURL)
 if err != nil {
  log.Fatalf("Failed to parse base URL: %v", err)
 }

 // Create a new Guacamole client
 client, err := guacamole.NewClient(parsedURL.String())
 if err != nil {
  log.Fatalf("Failed to create client: %v", err)
 }

 //Hardcoded Credentials
 username := "newuser"
 password := "password123"

 //Authenticate and get the auth token
 token, err := authenticate(client, username, password)
 if err != nil {
  log.Fatalf("Authentication failed: %v", err)
 }
 fmt.Printf("Authentication token: %s\n", *token.AuthToken)

 //Create a new User
 err = createUser(client, token.AuthToken)
 if err != nil {
  log.Fatalf("User creation failed: %v", err)
 }

 fmt.Println("User created successfully!")
}

// Authenticate with the Guacamole API and return the authentication token.
func authenticate(client *guacamole.Client, username, password string) (*guacamole.Authentication, error) {
 authParams := url.Values{}
 authParams.Set("username", username)
 authParams.Set("password", password)

 resp, err := client.AuthenticationLoginWithResponse(context.Background(), &authParams)
 if err != nil {
  return nil, fmt.Errorf("authentication failed: %w", err)
 }

 if resp.StatusCode() != http.StatusOK {
  return nil, fmt.Errorf("authentication failed with status code: %d", resp.StatusCode())
 }

 if resp.JSON200 == nil {
  return nil, fmt.Errorf("authentication response body is nil")
 }

 return resp.JSON200, nil
}

// createUser creates a new user in Guacamole.
func createUser(client *guacamole.Client, authToken *string) error {
 //Define the user to be created
 newUser := guacamole.User{
  Username:  "testuser",
  Password:  &guacamole.Password{Value: "testpassword"}, // Password now a pointer
  //Other optional fields can be set here
 }

 //Prepare the authentication context with the auth token
 ctx := context.WithValue(context.Background(), guacamole.GuacamoleTokenAuthScopes, []string{*authToken})

 // Define headers
 headers := map[string]string{"Content-Type": "application/json"}

 //Convert newUser to JSON
 requestBody, err := json.Marshal(newUser)
 if err != nil {
  return fmt.Errorf("failed to marshal user data: %w", err)
 }

 //Invoke the API to create the new user
 resp, err := client.CreateUserWithResponse(ctx, guacamole.CreateUserJSONRequestBody(requestBody), func(ctx context.Context, req *http.Request) error {
  for key, value := range headers {
   req.Header.Set(key, value)
  }
  req.Header.Set("Guacamole-Token", *authToken)
  return nil
 })
 if err != nil {
  return fmt.Errorf("failed to create user: %w", err)
 }

 if resp.StatusCode() != http.StatusCreated {
  bodyBytes, _ := io.ReadAll(resp.Body)
  bodyString := string(bodyBytes)
  return fmt.Errorf("failed to create user, status code: %d, body: %s", resp.StatusCode(), bodyString)
 }

 return nil
}

