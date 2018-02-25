package xray

type exception struct {
	// ID – A 64-bit identifier for the exception, unique among segments in the same trace,
	// in 16 hexadecimal digits.
	ID string `json:"id"`

	// Message – The exception message.
	Message string `json:"message,omitempty"`
}

// cause - A cause can be either a 16 character exception ID or an object with the following fields:
type errCause struct {
	// WorkingDirectory – The full path of the working directory when the exception occurred.
	WorkingDirectory string `json:"working_directory"`

	// Exceptions - The array of exception objects.
	Exceptions []exception `json:"exceptions,omitempty"`
}
