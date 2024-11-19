package converter

type ImageConverter interface {
	Shutdowner
	ToWebp(from string, to string, conf ConversionConfig) error
}

type Shutdowner interface {
	Shutdown()
}
