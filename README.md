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
go get https://github.com/elimity-com/go-scim-2.git
```

## Contributing
We are happy to review pull requests, 
but please first discuss the change you wish to make via issue, email, 
or any other method with the owners of this repository before making a change.

If you’d like to propose a change please ensure the following:
- all already existing tests are passing
- you have written tests that cover the code you are making
- there is documentation for at least all public functions you have added
- your changes are compliant with SCIM v2.0 (released as RFC7642, RFC7643 and RFC7644 under IETF)
