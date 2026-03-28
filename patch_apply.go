package scim

import (
	"fmt"

	scimErrors "github.com/elimity-com/scim/errors"
	internal "github.com/elimity-com/scim/filter"
	"github.com/elimity-com/scim/schema"
	filter "github.com/scim2/filter-parser/v2"
)

func applyToMatching(list []interface{}, t *valueExprTarget, value interface{}, replace bool) ([]interface{}, bool, error) {
	result := make([]interface{}, len(list))
	copy(result, list)

	matched := false
	matchedIndices := make(map[int]bool)
	for i, elem := range result {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		if !matchesExpression(m, t.expr, t.attr, t.refSchema) {
			continue
		}
		matched = true
		matchedIndices[i] = true
		if t.subAttrName != "" {
			updated := copyMap(m)
			updated[t.subAttrName] = value
			result[i] = updated
		} else if replace {
			if vm, ok := value.(map[string]interface{}); ok {
				updated := copyMap(m)
				for k, v := range vm {
					updated[k] = v
				}
				result[i] = updated
			} else {
				result[i] = value
			}
		} else {
			// Add to matching element.
			if vm, ok := value.(map[string]interface{}); ok {
				updated := copyMap(m)
				for k, v := range vm {
					if v != nil {
						updated[k] = v
					}
				}
				result[i] = updated
			} else {
				result[i] = value
			}
		}
	}

	if t.attr.HasPrimarySubAttr() {
		clearOtherPrimaries(result, matchedIndices)
	}

	return result, matched, nil
}

func attrFromTarget(target interface{}) schema.CoreAttribute {
	switch t := target.(type) {
	case *attributeTarget:
		return t.attr
	case *subAttributeTarget:
		return t.attr
	case *valueExprTarget:
		return t.attr
	}
	panic("unknown target type")
}

func checkMultiValuedConstraints(attrs ResourceAttributes, s schema.Schema, extensions []schema.Schema) error {
	allAttrs := make([]schema.CoreAttribute, 0, len(s.Attributes))
	allAttrs = append(allAttrs, s.Attributes...)
	for _, ext := range extensions {
		allAttrs = append(allAttrs, ext.Attributes...)
	}
	for _, attr := range allAttrs {
		if !attr.MultiValued() || !attr.HasSubAttributes() {
			continue
		}
		val, ok := attrs[attr.Name()]
		if !ok {
			continue
		}
		list, ok := val.([]interface{})
		if !ok {
			continue
		}
		if attr.HasTypeAndValueSubAttrs() && schema.HasDuplicateTypeValuePairs(list) {
			return scimErrors.ScimErrorInvalidValue
		}
		if attr.HasPrimarySubAttr() {
			clearDuplicatePrimary(list)
		}
	}
	return nil
}

// checkMutability validates that the operation is compatible with the
// attribute's mutability. Returns a mutability ScimError if not.
func checkMutability(op string, attr schema.CoreAttribute, exists bool) error {
	switch attr.Mutability() {
	case "readOnly":
		return scimErrors.ScimErrorMutability
	case "immutable":
		if op == PatchOperationAdd && !exists {
			return nil
		}
		return scimErrors.ScimErrorMutability
	}
	return nil
}

// clearDuplicatePrimary ensures at most one element has primary set to true.
// RFC 7644 Section 3.5.2: the server SHALL set the value of the existing
// "primary" attribute to false when a new primary value is added via PATCH.
// The last element with primary: true wins.
func clearDuplicatePrimary(list []interface{}) {
	lastPrimary := -1
	for i, elem := range list {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		if p, ok := m["primary"].(bool); ok && p {
			lastPrimary = i
		}
	}
	if lastPrimary < 0 {
		return
	}
	for i, elem := range list {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		if p, ok := m["primary"].(bool); ok && p && i != lastPrimary {
			m["primary"] = false
		}
	}
}

// clearOtherPrimaries clears primary on all elements outside modifiedIndices
// when any modified element has primary set to true. Among modified elements,
// only the last one with primary=true is kept.
func clearOtherPrimaries(list []interface{}, modifiedIndices map[int]bool) {
	lastModifiedPrimary := -1
	for i, elem := range list {
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		p, _ := m["primary"].(bool)
		if p && modifiedIndices[i] {
			lastModifiedPrimary = i
		}
	}
	if lastModifiedPrimary < 0 {
		return
	}
	for i, elem := range list {
		if i == lastModifiedPrimary {
			continue
		}
		m, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		if p, _ := m["primary"].(bool); p {
			m["primary"] = false
		}
	}
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func matchesExpression(element map[string]interface{}, expr filter.Expression, attr schema.CoreAttribute, refSchema schema.Schema) bool {
	v := internal.NewFilterValidator(expr, schema.Schema{
		ID:         refSchema.ID,
		Attributes: attr.SubAttributes(),
	})
	return v.PassesFilter(element) == nil
}

func mergeAdd(existing, value interface{}, attr schema.CoreAttribute) (interface{}, error) {
	if attr.MultiValued() {
		existingList, ok := existing.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected multi-valued attribute to be a list")
		}
		switch v := value.(type) {
		case []interface{}:
			return append(existingList, v...), nil
		default:
			return append(existingList, v), nil
		}
	}

	if attr.HasSubAttributes() {
		existingMap, ok := existing.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected complex attribute to be a map")
		}
		valueMap, ok := value.(map[string]interface{})
		if !ok {
			// Not a map, just replace.
			return value, nil
		}
		merged := copyMap(existingMap)
		for k, v := range valueMap {
			if v != nil {
				merged[k] = v
			}
		}
		return merged, nil
	}

	// Simple single-valued: replace.
	return value, nil
}

func removeMatching(list []interface{}, t *valueExprTarget) ([]interface{}, error) {
	var result []interface{}
	for _, elem := range list {
		m, ok := elem.(map[string]interface{})
		if !ok {
			result = append(result, elem)
			continue
		}
		if !matchesExpression(m, t.expr, t.attr, t.refSchema) {
			result = append(result, elem)
			continue
		}
		if t.subAttrName != "" {
			updated := copyMap(m)
			delete(updated, t.subAttrName)
			result = append(result, updated)
		}
		// If no sub-attribute, the matching element is removed entirely.
	}
	return result, nil
}

func resolveTarget(path *filter.Path, s schema.Schema, extensions []schema.Schema) (string, interface{}, error) {
	attrPath := path.AttributePath
	attrName := attrPath.AttributeName

	refSchema := s
	if uri := attrPath.URI(); uri != "" {
		found := false
		if uri == s.ID {
			found = true
		} else {
			for _, ext := range extensions {
				if uri == ext.ID {
					refSchema = ext
					found = true
					break
				}
			}
		}
		if !found {
			return "", nil, fmt.Errorf("unknown schema URI: %s", uri)
		}
	}

	attr, ok := refSchema.Attributes.ContainsAttribute(attrName)
	if !ok {
		return "", nil, fmt.Errorf("attribute not found: %s", attrName)
	}
	// Use the canonical name from the schema.
	attrName = attr.Name()

	// Extension attributes are stored under "schemaURI:attrName" in the resource.
	if uri := attrPath.URI(); uri != "" && uri != s.ID {
		attrName = uri + ":" + attrName
	}

	if path.ValueExpression != nil {
		subAttrName := path.SubAttributeName()
		return attrName, &valueExprTarget{
			attr:        attr,
			expr:        path.ValueExpression,
			subAttrName: subAttrName,
			refSchema:   refSchema,
		}, nil
	}

	if subAttrName := attrPath.SubAttributeName(); subAttrName != "" {
		// Resolve the canonical sub-attribute name.
		if attr.HasSubAttributes() {
			if sub, ok := attr.SubAttributes().ContainsAttribute(subAttrName); ok {
				subAttrName = sub.Name()
			}
		}
		return attrName, &subAttributeTarget{
			attr:        attr,
			subAttrName: subAttrName,
		}, nil
	}

	// Path-level sub-attribute (e.g., members[...].displayName has SubAttribute on Path).
	if subAttrName := path.SubAttributeName(); subAttrName != "" {
		if attr.HasSubAttributes() {
			if sub, ok := attr.SubAttributes().ContainsAttribute(subAttrName); ok {
				subAttrName = sub.Name()
			}
		}
		return attrName, &subAttributeTarget{
			attr:        attr,
			subAttrName: subAttrName,
		}, nil
	}

	return attrName, &attributeTarget{attr: attr}, nil
}

// ApplyPatch applies the given patch operations to the resource attributes.
// The schema and extensions are used to resolve attribute paths and validate
// value expressions. The operations are applied in order. The returned
// attributes are a modified copy of the input.
func ApplyPatch(attrs ResourceAttributes, ops []PatchOperation, s schema.Schema, extensions ...schema.Schema) (ResourceAttributes, error) {
	result := copyMap(attrs)
	for _, op := range ops {
		var err error
		result, err = applyOperation(result, op, s, extensions)
		if err != nil {
			return nil, err
		}
	}
	if err := checkMultiValuedConstraints(result, s, extensions); err != nil {
		return nil, err
	}
	return result, nil
}

func applyAdd(attrs ResourceAttributes, op PatchOperation, s schema.Schema, extensions []schema.Schema) (ResourceAttributes, error) {
	if op.Path == nil {
		return applyAddRoot(attrs, op.Value)
	}

	attrName, target, err := resolveTarget(op.Path, s, extensions)
	if err != nil {
		return nil, err
	}

	_, exists := attrs[attrName]
	if err := checkMutability(op.Op, attrFromTarget(target), exists); err != nil {
		return nil, err
	}

	switch t := target.(type) {
	case *attributeTarget:
		existing, exists := attrs[attrName]
		if !exists {
			attrs[attrName] = op.Value
			return attrs, nil
		}
		merged, err := mergeAdd(existing, op.Value, t.attr)
		if err != nil {
			return nil, err
		}
		attrs[attrName] = merged
	case *subAttributeTarget:
		existing, exists := attrs[attrName]
		if !exists {
			attrs[attrName] = map[string]interface{}{
				t.subAttrName: op.Value,
			}
			return attrs, nil
		}
		m, ok := existing.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected complex attribute for %s", attrName)
		}
		m = copyMap(m)
		m[t.subAttrName] = op.Value
		attrs[attrName] = m
	case *valueExprTarget:
		existing, exists := attrs[attrName]
		if !exists {
			return attrs, nil
		}
		list, ok := existing.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected multi-valued attribute for %s", attrName)
		}
		list, _, err := applyToMatching(list, t, op.Value, false)
		if err != nil {
			return nil, err
		}
		attrs[attrName] = list
	}
	return attrs, nil
}

func applyAddRoot(attrs ResourceAttributes, value interface{}) (ResourceAttributes, error) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value must be a map when path is omitted")
	}
	for k, v := range m {
		if v == nil {
			continue
		}
		existing, exists := attrs[k]
		if !exists {
			attrs[k] = v
			continue
		}
		// For multi-valued attributes, append.
		if existingList, ok := existing.([]interface{}); ok {
			if newList, ok := v.([]interface{}); ok {
				attrs[k] = append(existingList, newList...)
				continue
			}
		}
		// For complex attributes, merge.
		if existingMap, ok := existing.(map[string]interface{}); ok {
			if newMap, ok := v.(map[string]interface{}); ok {
				merged := copyMap(existingMap)
				for mk, mv := range newMap {
					if mv != nil {
						merged[mk] = mv
					}
				}
				attrs[k] = merged
				continue
			}
		}
		// Otherwise replace.
		attrs[k] = v
	}
	return attrs, nil
}

func applyOperation(attrs ResourceAttributes, op PatchOperation, s schema.Schema, extensions []schema.Schema) (ResourceAttributes, error) {
	switch op.Op {
	case PatchOperationAdd:
		return applyAdd(attrs, op, s, extensions)
	case PatchOperationReplace:
		return applyReplace(attrs, op, s, extensions)
	case PatchOperationRemove:
		return applyRemove(attrs, op, s, extensions)
	default:
		return nil, fmt.Errorf("unknown patch operation: %s", op.Op)
	}
}

func applyRemove(attrs ResourceAttributes, op PatchOperation, s schema.Schema, extensions []schema.Schema) (ResourceAttributes, error) {
	if op.Path == nil {
		return nil, scimErrors.ScimErrorNoTarget
	}

	attrName, target, err := resolveTarget(op.Path, s, extensions)
	if err != nil {
		return nil, err
	}

	_, exists := attrs[attrName]
	if err := checkMutability(op.Op, attrFromTarget(target), exists); err != nil {
		return nil, err
	}

	switch t := target.(type) {
	case *attributeTarget:
		if t.attr.Required() {
			return nil, scimErrors.ScimErrorInvalidValue
		}
		delete(attrs, attrName)
	case *subAttributeTarget:
		existing, exists := attrs[attrName]
		if !exists {
			return attrs, nil
		}
		m, ok := existing.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected complex attribute for %s", attrName)
		}
		m = copyMap(m)
		delete(m, t.subAttrName)
		if len(m) == 0 {
			delete(attrs, attrName)
		} else {
			attrs[attrName] = m
		}
	case *valueExprTarget:
		existing, exists := attrs[attrName]
		if !exists {
			return attrs, nil
		}
		list, ok := existing.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected multi-valued attribute for %s", attrName)
		}
		list, err := removeMatching(list, t)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			delete(attrs, attrName)
		} else {
			attrs[attrName] = list
		}
	}
	return attrs, nil
}

func applyReplace(attrs ResourceAttributes, op PatchOperation, s schema.Schema, extensions []schema.Schema) (ResourceAttributes, error) {
	if op.Path == nil {
		return applyReplaceRoot(attrs, op.Value)
	}

	attrName, target, err := resolveTarget(op.Path, s, extensions)
	if err != nil {
		return nil, err
	}

	_, exists := attrs[attrName]
	if err := checkMutability(op.Op, attrFromTarget(target), exists); err != nil {
		return nil, err
	}

	switch t := target.(type) {
	case *attributeTarget:
		// RFC 7644 Section 3.5.2.3: if the target location path specifies
		// an attribute that does not exist, the service provider SHALL
		// treat the operation as an "add".
		if _, exists := attrs[attrName]; !exists {
			return applyAdd(attrs, op, s, extensions)
		}
		// RFC 7644 Section 3.5.2.3: "If the target location specifies a
		// complex attribute, a set of sub-attributes SHALL be specified in
		// the 'value' parameter, which replaces any existing values or adds
		// where an attribute did not previously exist. Sub-attributes that
		// are not specified in the 'value' parameter are left unchanged."
		if t.attr.HasSubAttributes() && !t.attr.MultiValued() {
			existingMap, ok := attrs[attrName].(map[string]interface{})
			if ok {
				valueMap, ok := op.Value.(map[string]interface{})
				if ok {
					merged := copyMap(existingMap)
					for k, v := range valueMap {
						merged[k] = v
					}
					attrs[attrName] = merged
					return attrs, nil
				}
			}
		}
		attrs[attrName] = op.Value
	case *subAttributeTarget:
		existing, exists := attrs[attrName]
		if !exists {
			attrs[attrName] = map[string]interface{}{
				t.subAttrName: op.Value,
			}
			return attrs, nil
		}
		m, ok := existing.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected complex attribute for %s", attrName)
		}
		m = copyMap(m)
		m[t.subAttrName] = op.Value
		attrs[attrName] = m
	case *valueExprTarget:
		// RFC 7644 Section 3.5.2.3: if the target location is a multi-valued
		// attribute for which a value selection filter ("valuePath") has been
		// supplied and no record match was made, the service provider SHALL
		// indicate failure by returning HTTP status code 400 and a "scimType"
		// error code of "noTarget".
		existing, exists := attrs[attrName]
		if !exists {
			return nil, scimErrors.ScimErrorNoTarget
		}
		list, ok := existing.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected multi-valued attribute for %s", attrName)
		}
		list, matched, err := applyToMatching(list, t, op.Value, true)
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, scimErrors.ScimErrorNoTarget
		}
		attrs[attrName] = list
	}
	return attrs, nil
}

func applyReplaceRoot(attrs ResourceAttributes, value interface{}) (ResourceAttributes, error) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value must be a map when path is omitted")
	}
	for k, v := range m {
		if v == nil {
			delete(attrs, k)
			continue
		}
		attrs[k] = v
	}
	return attrs, nil
}

// target types returned by resolveTarget.
type attributeTarget struct {
	attr schema.CoreAttribute
}

type subAttributeTarget struct {
	attr        schema.CoreAttribute
	subAttrName string
}

type valueExprTarget struct {
	attr        schema.CoreAttribute
	expr        filter.Expression
	subAttrName string
	refSchema   schema.Schema
}
