package logger

func Prepare() {
	NewLogger(
		WithLevel("DEBUG"),
		WithAddSource(false),
		WithIsJSON(true),
		WithSetDefault(true))
}
