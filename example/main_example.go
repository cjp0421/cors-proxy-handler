package main

// Import only what we truly need for a minimal, clean Lambda handler.
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// handler is the core function that AWS Lambda runs for each request.
//
// The goal: absolutely minimal proxy logic with clean errors and CORS.
func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 1. Get the API key from environment variables.
	//    You will set API_KEY in the Lambda console.
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		// Missing API key = misconfigured server, not the user's fault.
		return serverError(fmt.Errorf("missing API key"))
	}

	// 2. Build the upstream API URL.
	//    Replace this example URL with the real API you want to call.
	upstreamURL := fmt.Sprintf("https://example.com/data?apikey=%s", apiKey)

	// 3. Make the HTTP request to the upstream API.
	resp, err := http.Get(upstreamURL)
	if err != nil {
		// The upstream API could not be reached at all.
		return serverError(fmt.Errorf("failed to reach upstream API: %w", err))
	}
	defer resp.Body.Close()

	// 4. Read the upstream API response.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return serverError(fmt.Errorf("failed to read upstream response: %w", err))
	}

	// 5. Return the upstream response directly to the browser.
	//    We forward the upstream status code, body, and add required CORS headers.
	return events.APIGatewayProxyResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		Headers:    corsHeaders(),
	}, nil
}

// clientError returns a 400-level error (bad input from the caller).
// Useful if you want to validate query params, paths, etc.
func clientError(msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       fmt.Sprintf(`{"error": "%s"}`, msg),
		Headers:    corsHeaders(),
	}, nil
}

// serverError returns a 500-level internal error.
// We use this for any failure inside the Lambda or bad upstream behavior.
func serverError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		Headers:    corsHeaders(),
	}, nil
}

// corsHeaders defines minimal required CORS headers so browsers can call the Lambda.
func corsHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": "*",
	}
}

// main tells AWS to use the handler above when the Lambda is invoked.
func main() {
	lambda.Start(handler)
}
