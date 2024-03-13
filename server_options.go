package scim

type ServerOption func(*Server)

// WithServiceProviderConfig sets the service provider config for the server.
func WithServiceProviderConfig(config ServiceProviderConfig) ServerOption {
	return func(s *Server) {
		s.config = config
	}
}

// WithResourceTypes sets the resource types for the server.
func WithResourceTypes(resourceTypes []ResourceType) ServerOption {
	return func(s *Server) {
		s.resourceTypes = resourceTypes
	}
}

// WithLogger sets the logger for the server.
func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		if logger != nil {
			s.log = logger
		}
	}
}
