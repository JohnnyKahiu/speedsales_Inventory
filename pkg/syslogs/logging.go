package syslogs

import (
	"os"
)

// LogRawRequest appends the provided request data to a raw file.
func LogRawRequest(data []byte) error {
	f, err := os.OpenFile(".requests.raw", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}
	_, err = f.Write([]byte("\n---\n"))
	return err
}
