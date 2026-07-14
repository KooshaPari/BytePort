package cloud

import (
	"io"
)

// maxProviderErrorBody is deliberately small: provider errors are diagnostic
// only and must not allow an untrusted response to allocate unbounded memory.
const maxProviderErrorBody = 64 * 1024

func readProviderErrorBody(r io.Reader) string {
	b, err := io.ReadAll(io.LimitReader(r, maxProviderErrorBody+1))
	if err != nil {
		return "<unable to read provider error body>"
	}
	if len(b) > maxProviderErrorBody {
		return string(b[:maxProviderErrorBody]) + "...[truncated]"
	}
	return string(b)
}
