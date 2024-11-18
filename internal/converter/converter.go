package converter

type ImageConverter interface {
	Shutdowner
	ToWebp(from string, to string) error
}

type Shutdowner interface {
	Shutdown()
}
