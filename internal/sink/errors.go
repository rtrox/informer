package sink

type InvalidConfigError struct{}

func (e InvalidConfigError) Error() string {
	return "Invalid Sink Config"
}
