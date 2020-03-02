package schema

import "github.com/elimity-com/scim/optional"

// CoreUserSchema returns the the default "User" Resource Schema.
func CoreUserSchema() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Unique identifier for the User, typically used by the user to directly authenticate to the service provider. Each User MUST include a non-empty userName value. This identifier MUST be unique across the service provider's entire set of Users. REQUIRED."),
				Name:        "userName",
				Required:    true,
				Uniqueness:  AttributeUniquenessServer(),
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("The components of the user's real name. Providers MAY return just the full name as a single string in the formatted sub-attribute, or they MAY return just the individual component attributes using the other sub-attributes, or they MAY return both. If both variants are returned, they SHOULD be describing the same name, with the formatted name indicating how the component attributes should be combined."),
				Name:        "name",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full name, including all middle names, titles, and suffixes as appropriate, formatted for display (e.g., 'Ms. Barbara J Jensen, III')."),
						Name:        "formatted",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The family name of the User, or last name in most Western languages (e.g., 'Jensen' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "familyName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The given name of the User, or first name in most Western languages (e.g., 'Barbara' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "givenName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The middle name(s) of the User (e.g., 'Jane' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "middleName",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The honorific prefix(es) of the User, or title in most Western languages (e.g., 'Ms.' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificPrefix",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The honorific suffix(es) of the User, or suffix in most Western languages (e.g., 'III' given the full name 'Ms. Barbara J Jensen, III')."),
						Name:        "honorificSuffix",
					}),
				},
			}),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The name of the User, suitable for display to end-users. The name SHOULD be the full name of the User being described, if known."),
				Name:        "displayName",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The casual way to address the user in real life, e.g., 'Bob' or 'Bobby' instead of 'Robert'. This attribute SHOULD NOT be used to represent a User's username (e.g., 'bjensen' or 'mpepperidge')."),
				Name:        "nickName",
			})),
			SimpleCoreAttribute(SimpleReferenceParams(ReferenceParams{
				Description:    optional.NewString("A fully qualified URL pointing to a page representing the User's online profile."),
				Name:           "profileUrl",
				ReferenceTypes: []AttributeReferenceType{AttributeReferenceTypeExternal},
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The user's title, such as \"Vice President.\""),
				Name:        "title",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Used to identify the relationship between the organization and the user. Typical values used might be 'Contractor', 'Employee', 'Intern', 'Temp', 'External', and 'Unknown', but any value may be used."),
				Name:        "userType",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Indicates the User's preferred written or spoken language. Generally used for selecting a localized user interface; e.g., 'en_US' specifies the language English and country US."),
				Name:        "preferredLanguage",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Used to indicate the User's default location for purposes of localizing items such as currency, date time format, or numerical representations."),
				Name:        "locale",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The User's time zone in the 'Olson' time zone database format, e.g., 'America/Los_Angeles'."),
				Name:        "timezone",
			})),
			SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
				Description: optional.NewString("A Boolean value indicating the User's administrative status."),
				Name:        "active",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("The User's cleartext password. This attribute is intended to be used as a means to specify an initial password when creating a new User or to reset an existing User's password."),
				Mutability:  AttributeMutabilityWriteOnly(),
				Name:        "password",
				Returned:    AttributeReturnedNever(),
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
				MultiValued: true,
				Name:        "emails",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Email addresses for the user. The value SHOULD be canonicalized by the service provider, e.g., 'bjensen@example.com' instead of 'bjensen@EXAMPLE.COM'. Canonical type values of 'work', 'home', and 'other'."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred mailing address or primary email address. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Phone numbers for the User. The value SHOULD be canonicalized by the service provider according to the format specified in RFC 3966, e.g., 'tel:+1-201-555-0123'. Canonical type values of 'work', 'home', 'mobile', 'fax', 'pager', and 'other'."),
				MultiValued: true,
				Name:        "phoneNumbers",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Phone number of the User."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "mobile", "fax", "pager", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work', 'home', 'mobile'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred phone number or primary phone number. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("Instant messaging addresses for the User."),
				MultiValued: true,
				Name:        "ims",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Instant messaging address for the User."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"aim", "gtalk", "icq", "xmpp", "msn", "skype", "qq", "yahoo"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'aim', 'gtalk', 'xmpp'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred messenger or primary messenger. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("URLs of photos of the User."),
				MultiValued: true,
				Name:        "photos",
				SubAttributes: []SimpleParams{
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("URL of a photo of the User."),
						Name:           "value",
						ReferenceTypes: []AttributeReferenceType{AttributeReferenceTypeExternal},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"photo", "thumbnail"},
						Description:     optional.NewString("A label indicating the attribute's function, i.e., 'photo' or 'thumbnail'."),
						Name:            "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute, e.g., the preferred photo or thumbnail. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A physical mailing address for this User. Canonical type values of 'work', 'home', and 'other'. This attribute is a complex type with the following sub-attributes."),
				MultiValued: true,
				Name:        "addresses",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full mailing address, formatted for display or use with a mailing label. This attribute MAY contain newlines."),
						Name:        "formatted",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The full street address component, which may include house number, street name, P.O. box, and multi-line extended street address information. This attribute MAY contain newlines."),
						Name:        "streetAddress",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The city or locality component."),
						Name:        "locality",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The state or region component."),
						Name:        "region",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The zip code or postal code component."),
						Name:        "postalCode",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The country name component."),
						Name:        "country",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"work", "home", "other"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'work' or 'home'."),
						Name:            "type",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of groups to which the user belongs, either through direct membership, through nested groups, or dynamically calculated."),
				MultiValued: true,
				Mutability:  AttributeMutabilityReadOnly(),
				Name:        "groups",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The identifier of the User's group."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI of the corresponding 'Group' resource to which the user belongs."),
						Mutability:     AttributeMutabilityReadOnly(),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User", "Group"},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"direct", "indirect"},
						Description:     optional.NewString("A label indicating the attribute's function, e.g., 'direct' or 'indirect'."),
						Mutability:      AttributeMutabilityReadOnly(),
						Name:            "type",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of entitlements for the User that represent a thing the User has."),
				MultiValued: true,
				Name:        "entitlements",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The value of an entitlement."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of roles for the User that collectively represent who the User is, e.g., 'Student', 'Faculty'."),
				MultiValued: true,
				Name:        "roles",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The value of a role."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of certificates issued to the User."),
				MultiValued: true,
				Name:        "x509Certificates",
				SubAttributes: []SimpleParams{
					SimpleBinaryParams(BinaryParams{
						Description: optional.NewString("The value of an X.509 certificate."),
						Name:        "value",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A human-readable name, primarily used for display purposes. READ-ONLY."),
						Name:        "display",
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("A label indicating the attribute's function."),
						Name:        "type",
					}),
					SimpleBooleanParams(BooleanParams{
						Description: optional.NewString("A Boolean value indicating the 'primary' or preferred attribute value for this attribute. The primary attribute value 'true' MUST appear no more than once."),
						Name:        "primary",
					}),
				},
			}),
		},
		Description: optional.NewString("User Account"),
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
	}
}

// CoreGroupSchema returns the the default "Group" Resource Schema.
func CoreGroupSchema() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("A human-readable name for the Group. REQUIRED."),
				Name:        "displayName",
				Required:    true,
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("A list of members of the Group."),
				MultiValued: true,
				Name:        "members",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("Identifier of the member of this Group."),
						Mutability:  AttributeMutabilityImmutable(),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI corresponding to a SCIM resource that is a member of this Group."),
						Mutability:     AttributeMutabilityImmutable(),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User", "Group"},
					}),
					SimpleStringParams(StringParams{
						CanonicalValues: []string{"User", "Group"},
						Description:     optional.NewString("A label indicating the type of resource, e.g., 'User' or 'Group'."),
						Mutability:      AttributeMutabilityImmutable(),
						Name:            "type",
					}),
				},
			}),
		},
		Description: optional.NewString("Group"),
		ID:          "urn:ietf:params:scim:schemas:core:2.0:Group",
		Name:        optional.NewString("Group"),
	}
}

// ExtensionEnterpriseUser returns the the default Enterprise User Schema Extension.
func ExtensionEnterpriseUser() Schema {
	return Schema{
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Numeric or alphanumeric identifier assigned to a person, typically based on order of hire or association with an organization."),
				Name:        "employeeNumber",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a cost center."),
				Name:        "costCenter",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of an organization."),
				Name:        "organization",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a division."),
				Name:        "division",
			})),
			SimpleCoreAttribute(SimpleStringParams(StringParams{
				Description: optional.NewString("Identifies the name of a department."),
				Name:        "department",
			})),
			ComplexCoreAttribute(ComplexParams{
				Description: optional.NewString("The User's manager. A complex type that optionally allows service providers to represent organizational hierarchy by referencing the 'id' attribute of another User."),
				Name:        "manager",
				SubAttributes: []SimpleParams{
					SimpleStringParams(StringParams{
						Description: optional.NewString("The id of the SCIM resource representing the User's manager. REQUIRED."),
						Name:        "value",
					}),
					SimpleReferenceParams(ReferenceParams{
						Description:    optional.NewString("The URI of the SCIM resource representing the User's manager. REQUIRED."),
						Name:           "$ref",
						ReferenceTypes: []AttributeReferenceType{"User"},
					}),
					SimpleStringParams(StringParams{
						Description: optional.NewString("The displayName of the User's manager. OPTIONAL and READ-ONLY."),
						Mutability:  AttributeMutabilityReadOnly(),
						Name:        "displayName",
					}),
				},
			}),
		},
		Description: optional.NewString("Enterprise User"),
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("Enterprise User"),
	}
}
