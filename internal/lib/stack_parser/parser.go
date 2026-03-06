package stack_parser

type StackParser interface {
	Parse([]byte, any) ([]byte, error)
}
