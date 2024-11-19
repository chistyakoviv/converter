package converter

type ImageConverter interface {
	Shutdowner
	ToWebp(from string, to string, conf ConversionConfig) error
}

type VideoConverter interface {
	Shutdowner
	ToWebm(from string, to string, conf ConversionConfig) error
}

type Shutdowner interface {
	Shutdown()
}
