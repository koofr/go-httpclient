package httpclient

import (
	"fmt"
	"net/http"
)

type InvalidStatusError struct {
	Expected []int
	Got      int
	Headers  http.Header
	Content  string
}

func (e InvalidStatusError) Error() string {
	return fmt.Sprintf("Invalid response status! Got %d, expected %d; headers: %s, content: %s", e.Got, e.Expected, e.Headers, e.Content)
}
