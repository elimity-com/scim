# Internal Filter Package

## Validator
It is possible to create a new (filter) validator by either passing a parsed `filter.Expression` or a string that still
needs to be parsed. The latter is **recommended**, but can result in an error if the filter is not valid.

It is **recommended** to always invoke `validator.Validate()` to validate whether the filter matches the given schemas.

The method `validatorPassesFilter(resource map[string]interface{})` checks whether the given resource matches the filter.
There are **specific** value types expected for this validation to pass. If another type is encountered a `errors.ScimError`
gets returned, a `fmt.Errorf` gets returned if the validation fails.

### Overview of Expected Types
|SCIM Type         |Go Type |
|---               |--- |
|binary            |string |
|dateTime          |string, xsd:dateTime |
|reference, string |string |
|boolean           |boolean |
|decimal           |float64, float32, int64, ..., int |
|integer           |int64, int32, ..., int |
