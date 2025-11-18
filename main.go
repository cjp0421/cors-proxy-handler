package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// handler is the main Lambda entrypoint.
// It:
//   - reads ?id=<planet> from the query string (e.g., ?id=earth)
//   - calls the French Solar System API for that body
//   - passes your API key as a Bearer token in the Authorization header
//   - returns the JSON response to the browser with CORS headers.
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	// 1. Read API key from environment variable.
	//    You will configure API_KEY in the Lambda console.
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		// Missing key = server misconfiguration.
		return serverError(fmt.Errorf("missing API_KEY environment variable"))
	}

	// 2. Read the planet/body ID from the query string: ?id=earth
	//
	//    Example frontend call:
	//      fetch("https://<api-gw-url>/proxy?id=earth")
	//
	//    You can choose the param name; here we use "id".
	bodyID, ok := req.QueryStringParameters["id"]
	if !ok || bodyID == "" {
		// This is a client error: the caller didn't provide required input.
		return clientError("missing required query parameter 'id'")
	}

	// 3. Build the upstream URL for this specific body.
	//
	//    French Solar System API pattern:
	//      GET https://api.le-systeme-solaire.net/rest/bodies/{id}
	//
	//    Examples:
	//      /rest/bodies/earth
	//      /rest/bodies/mars
	//
	//    In production, we call the real French Solar System API.
	//    For tests, we can override this base URL using UPSTREAM_BASE_URL.
	baseURL := os.Getenv("UPSTREAM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.le-systeme-solaire.net/rest/bodies"
	}

	upstreamURL := fmt.Sprintf("%s/%s", baseURL, bodyID)

	// 4. Create a new HTTP request so we can attach headers (Authorization: Bearer <key>).
	reqUpstream, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL, nil)
	if err != nil {
		return serverError(fmt.Errorf("failed to create upstream request: %w", err))
	}

	// 5. Add Authorization header with Bearer token.
	//
	//    This is where your API key is actually used.
	reqUpstream.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 6. Perform the HTTP request with the default HTTP client.
	resp, err := http.DefaultClient.Do(reqUpstream)
	if err != nil {
		// We couldn't reach the upstream API at all.
		return serverError(fmt.Errorf("failed to reach solar system API: %w", err))
	}
	defer resp.Body.Close()

	// 7. Read the upstream response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return serverError(fmt.Errorf("failed to read solar system API response: %w", err))
	}

	// 8. Return the upstream response back to the browser.
	//
	//    - We forward the status code from the upstream API.
	//    - We return the raw JSON body as-is.
	//    - We attach CORS headers so your frontend can call this Lambda.
	return events.APIGatewayProxyResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		Headers:    corsHeaders(),
	}, nil
}

// clientError represents a 400-level response: the caller sent a bad request.
// Example: missing query parameters, invalid values, etc.
func clientError(msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       fmt.Sprintf(`{"error": "%s"}`, msg),
		Headers:    corsHeaders(),
	}, nil
}

// serverError represents a 500-level response: something went wrong on the server side.
// Example: missing API_KEY, network error calling upstream, JSON read failures, etc.
func serverError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		Headers:    corsHeaders(),
	}, nil
}

// corsHeaders returns the minimal CORS headers needed so the browser
// will accept the response from your Lambda when called via fetch().
func corsHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": "*",
	}
}

// main tells the AWS Lambda runtime to use the handler function above
// for each incoming invocation triggered by API Gateway.
func main() {
	lambda.Start(handler)
}
