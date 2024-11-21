package converter

type ImageConverter interface {
	Shutdowner
	Convert(from string, to string, conf ConversionConfig) error
}

type VideoConverter interface {
	Shutdowner
	Convert(from string, to string, conf ConversionConfig) error
}

type Shutdowner interface {
	Shutdown()
}
