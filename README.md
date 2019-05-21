![scim-logo](LOGO.png)

[![GoDoc](https://godoc.org/github.com/elimity.com/scim?status.svg)](https://godoc.org/github.com/elimity.com/scim)

This is an open source implementation of the [SCIM v2.0](http://www.simplecloud.info/#Specification) specification for use in Golang.
SCIM defines a flexible schema mechanism and REST API for managing identity data.
The goal is to reduce the complexity of user management operations by providing patterns for exchanging schemas using HTTP.

In this implementation it is easy to add *custom* schemas and extensions whom are validated at the initialization of the server.
Corresponding with their resource type, incoming resources will be *validated* by the supported schemas before being 
passed on to their callbacks.

The following features are supported:
- GET for `/Schemas`, `/ServiceProviderConfig` and `/ResourceTypes`
- CRUD (POST/GET/PUT and DELETE) for your own resource types (i.e. `/Users`, `/Groups`, `/Employees`, ...)

Other optional features such as patch, pagination, sorting, etc... are **not** supported in this version.

## Installation
Assuming you already have a (recent) version of Go installed, you can get the code with go get:
```
go get github.com/elimity-com/scim
```

## Usage
**!** errors are ignored for simplicity.
### 1. Create a service provider configuration.
[RFC Config](https://tools.ietf.org/html/rfc7643#section-5) |
[Example Config](https://tools.ietf.org/html/rfc7643#section-8.5)
```
// i.e. rawConfig, _ := ioutil.ReadFile("/path/to/config")
config, _ := scim.NewServiceProviderConfig(rawConfig)
```
**!** no additional features/operations are supported in this version.

### 2. Create all supported schemas and extensions.
[RFC Schema](https://tools.ietf.org/html/rfc7643#section-2) |
[User Schema](https://tools.ietf.org/html/rfc7643#section-4.1) |
[Group Schema](https://tools.ietf.org/html/rfc7643#section-4.2) |
[Extension Schema](https://tools.ietf.org/html/rfc7643#section-4.3)
```
schema, _ := scim.NewSchema(rawSchema)
extension, _ := scim.NewSchema(rawExtension)
```

### 3. Create all resource types and their callbacks.
[RFC Resource Type](https://tools.ietf.org/html/rfc7643#section-6) |
[Example Resource Type](https://tools.ietf.org/html/rfc7643#section-8.6)

#### 3.1 Callback (implementation of `ResourceHandler`)
[Simple In Memory Example](resource_handler_test.go)
```
var resourceHandler scim.ResourceHandler
// initialize w/ own implementation
```
**!** each resource type should have its own resource handler.

#### 3.2 Resource Type
```
resourceType, _ := scim.NewResourceType(rawResourceType, resourceHandler)
```
**!** make sure all schemas that are referenced are created in the previous step.

### 4. Create Server
```
server, _ := scim.NewServer(config, []scim.Schema{schema, extension}, []scim.ResourceType{resourceType})
```

### 5. Listen and Serve
```
log.Fatal(http.ListenAndServe(":8080", server))
```

## Contributing
We are happy to review pull requests, 
but please first discuss the change you wish to make via issue, email, 
or any other method with the owners of this repository before making a change.

If you’d like to propose a change please ensure the following:
- all checks of CircleCI are passing ([golangci-lint](https://github.com/golangci/golangci-lint): `goimports` and `golint`)
- all already existing tests are passing
- you have written tests that cover the code you are making
- there is documentation for at least all public functions you have added
- your changes are compliant with SCIM v2.0 (released as 
[RFC7642](https://tools.ietf.org/html/rfc7642), 
[RFC7643](https://tools.ietf.org/html/rfc7643) and 
[RFC7644](https://tools.ietf.org/html/rfc7644) under [IETF](https://ietf.org/))
