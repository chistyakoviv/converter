package handlers

type Validator interface {
	Struct(s interface{}) error
}
