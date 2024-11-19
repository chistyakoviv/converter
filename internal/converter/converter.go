package converter

// TODO: rewrite to use common interface for both image and video
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
