package scim

type Server struct {
	schemas []Schema
}

func NewServer(schemas ...Schema) Server {
	return Server{
		schemas: schemas,
	}
}
