package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

const (
	// UserSchema is the URI for the User resource.
	UserSchema = "urn:ietf:params:scim:schemas:core:2.0:User"

	// GroupSchema is the URI for the Group resource.
	GroupSchema = "urn:ietf:params:scim:schemas:core:2.0:Group"
)

func cannotBePatched(op string, attr CoreAttribute) bool {
	return isImmutable(op, attr) || isReadOnly(attr)
}

func isImmutable(op string, attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityImmutable && (op == "replace" || op == "remove")
}

func isReadOnly(attr CoreAttribute) bool {
	return attr.mutability == attributeMutabilityReadOnly
}

// validateUnmarshalAttributeName validates an attribute name for JSON
// unmarshaling. It applies the same rules as the constructors but also
// accepts "$ref", which RFC 7643 Section 2.4 defines as a standard
// sub-attribute despite it violating the ABNF grammar.
func validateUnmarshalAttributeName(name string) error {
	if name == "$ref" {
		return nil
	}
	return validateAttributeName(name)
}

// Attributes represent a list of Core Attributes.
type Attributes []CoreAttribute

func unmarshalAttributes(rawAttrs []json.RawMessage) (Attributes, error) {
	attrs := make(Attributes, 0, len(rawAttrs))
	for _, raw := range rawAttrs {
		a, err := unmarshalCoreAttribute(raw)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	return attrs, nil
}

// ContainsAttribute checks whether the list of Core Attributes contains an attribute with the given name.
func (as Attributes) ContainsAttribute(name string) (CoreAttribute, bool) {
	for _, a := range as {
		if strings.EqualFold(name, a.name) {
			return a, true
		}
	}
	return CoreAttribute{}, false
}

func unmarshalCoreAttribute(data json.RawMessage) (CoreAttribute, error) {
	var raw struct {
		Name            string            `json:"name"`
		Type            string            `json:"type"`
		Description     optional.String   `json:"description"`
		MultiValued     bool              `json:"multiValued"`
		Required        bool              `json:"required"`
		CaseExact       bool              `json:"caseExact"`
		Mutability      string            `json:"mutability"`
		Returned        string            `json:"returned"`
		Uniqueness      string            `json:"uniqueness"`
		CanonicalValues []string          `json:"canonicalValues"`
		ReferenceTypes  []string          `json:"referenceTypes"`
		SubAttributes   []json.RawMessage `json:"subAttributes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return CoreAttribute{}, err
	}

	if err := validateUnmarshalAttributeName(raw.Name); err != nil {
		return CoreAttribute{}, err
	}

	typ, err := parseAttributeType(raw.Type)
	if err != nil {
		return CoreAttribute{}, err
	}

	mut, err := parseAttributeMutability(raw.Mutability)
	if err != nil {
		return CoreAttribute{}, err
	}

	ret, err := parseAttributeReturned(raw.Returned)
	if err != nil {
		return CoreAttribute{}, err
	}

	uniq, err := parseAttributeUniqueness(raw.Uniqueness)
	if err != nil {
		return CoreAttribute{}, err
	}

	if typ == attributeDataTypeComplex {
		subParams, err := unmarshalSimpleParams(raw.SubAttributes)
		if err != nil {
			return CoreAttribute{}, err
		}
		subAttrs, err := buildSubAttributes(subParams)
		if err != nil {
			return CoreAttribute{}, err
		}
		return CoreAttribute{
			description:   raw.Description,
			multiValued:   raw.MultiValued,
			mutability:    mut,
			name:          raw.Name,
			required:      raw.Required,
			returned:      ret,
			subAttributes: subAttrs,
			typ:           attributeDataTypeComplex,
			uniqueness:    uniq,
		}, nil
	}

	var refTypes []AttributeReferenceType
	for _, r := range raw.ReferenceTypes {
		refTypes = append(refTypes, AttributeReferenceType(r))
	}

	return CoreAttribute{
		canonicalValues: raw.CanonicalValues,
		caseExact:       raw.CaseExact,
		description:     raw.Description,
		multiValued:     raw.MultiValued,
		mutability:      mut,
		name:            raw.Name,
		referenceTypes:  refTypes,
		required:        raw.Required,
		returned:        ret,
		typ:             typ,
		uniqueness:      uniq,
	}, nil
}

// Schema is a collection of attribute definitions that describe the contents of an entire or partial resource.
type Schema struct {
	Attributes  Attributes
	Description optional.String
	ID          string
	Name        optional.String
}

// MarshalJSON converts the schema struct to its corresponding json representation.
func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ToMap())
}

// ToMap returns the map representation of a schema.
func (s Schema) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name.Value(),
		"description": s.Description.Value(),
		"attributes":  s.getRawAttributes(),
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"},
	}
}

// UnmarshalJSON parses a JSON-encoded schema into the Schema struct.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID          string            `json:"id"`
		Name        optional.String   `json:"name"`
		Description optional.String   `json:"description"`
		Attributes  []json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	attrs, err := unmarshalAttributes(raw.Attributes)
	if err != nil {
		return err
	}

	s.ID = raw.ID
	s.Name = raw.Name
	s.Description = raw.Description
	s.Attributes = attrs
	return nil
}

// Validate validates given resource based on the schema, including the
// "schemas" attribute. Does NOT validate mutability.
// NOTE: only used in POST and PUT requests where attributes MAY be (re)defined.
func (s Schema) Validate(resource interface{}) (map[string]interface{}, *errors.ScimError) {
	return s.validate(resource, false, true)
}

// ValidateExtension validates an extension resource without checking the
// "schemas" attribute, since extensions are nested under their schema ID
// and do not carry their own "schemas" array.
func (s Schema) ValidateExtension(resource interface{}) (map[string]interface{}, *errors.ScimError) {
	return s.validate(resource, false, false)
}

// ValidateMutability validates given resource based on the schema, including strict immutability checks.
func (s Schema) ValidateMutability(resource interface{}) (map[string]interface{}, *errors.ScimError) {
	return s.validate(resource, true, false)
}

// ValidatePatchOperation validates an individual operation and its related value.
func (s Schema) ValidatePatchOperation(operation string, operationValue map[string]interface{}, isExtension bool) *errors.ScimError {
	for k, v := range operationValue {
		var attr *CoreAttribute
		var scimErr *errors.ScimError

		for _, attribute := range s.Attributes {
			if strings.EqualFold(attribute.name, k) {
				attr = &attribute
				break
			}
			if isExtension && strings.EqualFold(s.ID+":"+attribute.name, k) {
				attr = &attribute
				break
			}
		}

		// Attribute does not exist in the schema, thus it is an invalid request.
		// Immutable attrs can only be added and Readonly attrs cannot be patched
		if attr == nil || cannotBePatched(operation, *attr) {
			return &errors.ScimErrorInvalidValue
		}

		// "remove" operations simply have to exist
		if operation != "remove" {
			_, scimErr = attr.validate(v)
		}

		if scimErr != nil {
			return scimErr
		}
	}

	return nil
}

// ValidatePatchOperationValue validates an individual operation and its related value.
func (s Schema) ValidatePatchOperationValue(operation string, operationValue map[string]interface{}) *errors.ScimError {
	return s.ValidatePatchOperation(operation, operationValue, false)
}

func (s Schema) getRawAttributes() []map[string]interface{} {
	attributes := make([]map[string]interface{}, len(s.Attributes))

	for i, a := range s.Attributes {
		attributes[i] = a.getRawAttributes()
	}

	return attributes
}

func (s Schema) validate(resource interface{}, checkMutability, checkSchemaID bool) (map[string]interface{}, *errors.ScimError) {
	core, ok := resource.(map[string]interface{})
	if !ok {
		return nil, &errors.ScimErrorInvalidSyntax
	}

	if checkSchemaID {
		if err := s.validateSchemaID(core); err != nil {
			return nil, err
		}
	}

	attributes := make(map[string]interface{})
	for _, attribute := range s.Attributes {
		var hit interface{}
		var found bool
		for k, v := range core {
			if strings.EqualFold(attribute.name, k) {
				// duplicate found
				if found {
					return nil, &errors.ScimErrorInvalidSyntax
				}
				found = true
				hit = v
			}
		}

		// An immutable attribute SHALL NOT be updated.
		if found && checkMutability &&
			attribute.mutability == attributeMutabilityImmutable {
			return nil, &errors.ScimErrorMutability
		}

		attr, scimErr := attribute.validate(hit)
		if scimErr != nil {
			return nil, scimErr
		}
		if attr != nil {
			attributes[attribute.name] = attr
		}
	}
	return attributes, nil
}

func (s Schema) validateSchemaID(resource map[string]interface{}) *errors.ScimError {
	resourceSchemas, present := resource["schemas"]
	if !present {
		return &errors.ScimErrorInvalidSyntax
	}

	resourceSchemasSlice, ok := resourceSchemas.([]interface{})
	if !ok {
		return &errors.ScimErrorInvalidSyntax
	}

	var schemaFound bool
	for _, v := range resourceSchemasSlice {
		if v == s.ID {
			schemaFound = true
			break
		}
	}
	if !schemaFound {
		return &errors.ScimErrorInvalidSyntax
	}

	return nil
}

func unmarshalSimpleParam(data json.RawMessage) (SimpleParams, error) {
	var raw struct {
		Name            string          `json:"name"`
		Type            string          `json:"type"`
		Description     optional.String `json:"description"`
		MultiValued     bool            `json:"multiValued"`
		Required        bool            `json:"required"`
		CaseExact       bool            `json:"caseExact"`
		Mutability      string          `json:"mutability"`
		Returned        string          `json:"returned"`
		Uniqueness      string          `json:"uniqueness"`
		CanonicalValues []string        `json:"canonicalValues"`
		ReferenceTypes  []string        `json:"referenceTypes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return SimpleParams{}, err
	}

	if err := validateUnmarshalAttributeName(raw.Name); err != nil {
		return SimpleParams{}, err
	}

	typ, err := parseAttributeType(raw.Type)
	if err != nil {
		return SimpleParams{}, err
	}

	mut, err := parseAttributeMutability(raw.Mutability)
	if err != nil {
		return SimpleParams{}, err
	}

	ret, err := parseAttributeReturned(raw.Returned)
	if err != nil {
		return SimpleParams{}, err
	}

	uniq, err := parseAttributeUniqueness(raw.Uniqueness)
	if err != nil {
		return SimpleParams{}, err
	}

	var refTypes []AttributeReferenceType
	for _, r := range raw.ReferenceTypes {
		refTypes = append(refTypes, AttributeReferenceType(r))
	}

	return SimpleParams{
		canonicalValues: raw.CanonicalValues,
		caseExact:       raw.CaseExact,
		description:     raw.Description,
		multiValued:     raw.MultiValued,
		mutability:      mut,
		name:            raw.Name,
		referenceTypes:  refTypes,
		required:        raw.Required,
		returned:        ret,
		typ:             typ,
		uniqueness:      uniq,
	}, nil
}

func unmarshalSimpleParams(rawAttrs []json.RawMessage) ([]SimpleParams, error) {
	params := make([]SimpleParams, 0, len(rawAttrs))
	for _, raw := range rawAttrs {
		p, err := unmarshalSimpleParam(raw)
		if err != nil {
			return nil, err
		}
		params = append(params, p)
	}
	return params, nil
}

func parseAttributeMutability(s string) (attributeMutability, error) {
	switch s {
	case "readWrite", "":
		return attributeMutabilityReadWrite, nil
	case "immutable":
		return attributeMutabilityImmutable, nil
	case "readOnly":
		return attributeMutabilityReadOnly, nil
	case "writeOnly":
		return attributeMutabilityWriteOnly, nil
	default:
		return 0, fmt.Errorf("unknown mutability: %q", s)
	}
}

func parseAttributeReturned(s string) (attributeReturned, error) {
	switch s {
	case "default", "":
		return attributeReturnedDefault, nil
	case "always":
		return attributeReturnedAlways, nil
	case "never":
		return attributeReturnedNever, nil
	case "request":
		return attributeReturnedRequest, nil
	default:
		return 0, fmt.Errorf("unknown returned: %q", s)
	}
}

func parseAttributeType(s string) (attributeType, error) {
	switch s {
	case "string":
		return attributeDataTypeString, nil
	case "boolean":
		return attributeDataTypeBoolean, nil
	case "decimal":
		return attributeDataTypeDecimal, nil
	case "integer":
		return attributeDataTypeInteger, nil
	case "dateTime":
		return attributeDataTypeDateTime, nil
	case "reference":
		return attributeDataTypeReference, nil
	case "complex":
		return attributeDataTypeComplex, nil
	case "binary":
		return attributeDataTypeBinary, nil
	default:
		return 0, fmt.Errorf("unknown attribute type: %q", s)
	}
}

func parseAttributeUniqueness(s string) (attributeUniqueness, error) {
	switch s {
	case "none", "":
		return attributeUniquenessNone, nil
	case "server":
		return attributeUniquenessServer, nil
	case "global":
		return attributeUniquenessGlobal, nil
	default:
		return 0, fmt.Errorf("unknown uniqueness: %q", s)
	}
}
