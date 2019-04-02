package scim

type Server struct {
	schemas []schema
}

func NewServer(schemas ...schema) Server {
	return Server{
		schemas: schemas,
	}
}
