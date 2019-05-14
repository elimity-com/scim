# Go SCIM 2
The System for Cross-domain Identity Management (SCIM) specification is designed to make managing user identities in 
cloud-based applications and services easier. The specification suite seeks to build upon experience with existing 
schemas and deployments, placing specific emphasis on simplicity of development and integration, while applying 
existing authentication, authorization, and privacy models. Its intent is to reduce the cost and complexity of user 
management operations by providing a common user schema and extension model, as well as binding documents to provide 
patterns for exchanging this schema using standard protocols. ([more...](http://www.simplecloud.info/))

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
config, _ := NewServiceProviderConfigFromFile("/path/to/config")
```
**!** no additional features/operations are supported in this version.

### 2. Create all supported schemas and extensions.
[RFC Schema](https://tools.ietf.org/html/rfc7643#section-2) |
[User Schema](https://tools.ietf.org/html/rfc7643#section-4.1) |
[Group Schema](https://tools.ietf.org/html/rfc7643#section-4.2) |
[Extension Schema](https://tools.ietf.org/html/rfc7643#section-4.3)
```
schema, _ := NewSchemaFromFile("/path/to/schema")
extension, _ := NewSchemaFromFile("/path/to/extension")
```

### 3. Create all resource types and their callbacks.
[RFC Resource Type](https://tools.ietf.org/html/rfc7643#section-6) |
[Example Resource Type](https://tools.ietf.org/html/rfc7643#section-8.6)

#### 3.1 Callback (implementation of `ResourceHandler`)
[Simple In Memory Example](resource_handler_test.go)
```
var resourceHandler ResourceHandler
// initialize w/ own implementation
```
**!** each resource type should have its own resource handler.

#### 3.2 Resource Type
```
resourceType, _ := NewResourceTypeFromFile("/path/to/resourceType", resourceHandler)
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
- all already existing tests are passing
- you have written tests that cover the code you are making
- there is documentation for at least all public functions you have added
- your changes are compliant with SCIM v2.0 (released as 
[RFC7642](https://tools.ietf.org/html/rfc7642), 
[RFC7643](https://tools.ietf.org/html/rfc7643) and 
[RFC7644](https://tools.ietf.org/html/rfc7644) under [IETF](https://ietf.org/))
