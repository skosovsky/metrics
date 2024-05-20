package logger

func simple() { //nolint:unused // example
	logger := NewLogger(WithSetDefault(true))

	logger.Info("Hello, World!")

	logger = NewLogger(WithLevel("DEBUG"), WithAddSource(false))
	logger.Debug("Hello, World!")

	logger = NewLogger(WithLevel("DEBUG"), WithAddSource(false), WithIsJSON(false))
	logger.Debug("Hello, World!")

	_ = NewLogger(WithAddSource(true))
	Default().Info("Hello, World!")

	logger = WithAttrs(logger, StringAttr("hello", "world"))
	logger.Info("OK")
}
