# **AWS Lambda Go Proxy Handler**
[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
![Status](https://img.shields.io/badge/status-production--ready-blue)
![Tests](https://img.shields.io/badge/tests-passing-brightgreen)

A minimal Go-based AWS Lambda handler used to proxy requests from a frontend application to an upstream API.  
The handler injects a Bearer token using a Lambda environment variable, forwards the upstream response, and adds CORS headers so it can be safely called from a browser.

This repository contains **two versions** of the handler:  
a **production version**, and a **simplified teaching example** used in documentation.

---

## **📁 Repository Structure**

```
/
├─ main.go                # Production handler used in AWS Lambda
├─ main_test.go           # Integration-style tests using httptest.Server
├─ example/
│   └─ main_example.go    # Minimal teaching/demo version of the handler
├─ build.sh               # Compile + package for Lambda
├─ go.mod / go.sum        # Go module files
└─ README.md
```

---

## **🚀 main.go — Production Handler**

`main.go` is the **real** Lambda function. It includes:

- Query parameter support (e.g., `?id=earth`)
- API key read from environment variables (`API_KEY`)
- `Authorization: Bearer <key>` header
- Clean 400/500 error helpers  
- CORS headers for browser use  
- Optional `UPSTREAM_BASE_URL` override for tests  
- Full integration into AWS Lambda → API Gateway

This is the version built, zipped, and uploaded using the included `build.sh` script.

---

## **📘 example/main_example.go — Minimal Teaching Example**

This file contains a **simplified handler** used for documentation.

It demonstrates only the essentials:

- Receiving a request from API Gateway  
- Calling an upstream API  
- Returning JSON + CORS  
- Basic error handling structure  

It does **not** include query param support, Bearer token injection, or real-world behavior.  
It’s a clean, minimal version used to teach the handler pattern before introducing the production implementation.

---

## **🧪 Tests**

`main_test.go` contains integration-style tests using Go’s `httptest.Server` to mock the upstream API.  
Tests verify:

- Successful proxy calls  
- Missing query parameters  
- Missing API key  
- Upstream 500 errors  
- Correct forwarding of status, body, and headers  
- Authorization header correctness  

Run all tests:

```sh
go test ./...
```

---

## **🔨 Build & Deploy**

Compile for AWS Lambda (Linux, amd64) and package:

```sh
./build.sh
```

This produces:

- `bootstrap` — the Lambda binary  
- `function.zip` — upload this to AWS Lambda  

---

## **📄 License**

MIT License — see `LICENSE` for details.