# Clinia HTTP Client

Clinia's HTTP Client that helps to make HTTP requests easier

### Usage

#### Get Clinia's HTTP client with default settings
The default client comes with a default timeout of 60 seconds
```go
client := NewHTTPClient()
```

#### Get Clinia's HTTP client with customizable settings
This client is helpful if you want settings to be customized, like setting different timeout, setting proxy, etc
```go
client := NewClientWithOptions(
    WithTimeout(10 * time.Second),
)
```

#### Make HTTP requests with the Clinia client
```go
ctx := context.Background()

client := NewClientWithOptions(
    WithTimeout(10 * time.Second),
)

response, err := client.MakeHTTPRequest(ctx, &Request{
	Method: "GET",
	URL: "http://example.com",
})
```
See the `Request` and `Response` struct for all the available fields

#### Get Go's http.Client with default settings
This client is useful if you do not want to use Clinia's HTTP client
```go
client := GetDefaultHTTPClient()
```