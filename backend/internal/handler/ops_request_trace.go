package handler

import (
	"bytes"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	opsResponseCaptureKey   = "ops_response_capture"
	opsResponseCaptureLimit = 0
)

type opsResponseCaptureWriter struct {
	gin.ResponseWriter
	limit int
	buf   bytes.Buffer
	total int
}

func newOpsResponseCaptureWriter(w gin.ResponseWriter, limit int) *opsResponseCaptureWriter {
	if limit <= 0 {
		limit = opsResponseCaptureLimit
	}
	return &opsResponseCaptureWriter{
		ResponseWriter: w,
		limit:          limit,
	}
}

func (w *opsResponseCaptureWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *opsResponseCaptureWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *opsResponseCaptureWriter) capture(data []byte) {
	if len(data) == 0 {
		return
	}
	w.total += len(data)
	if w.limit > 0 {
		if w.buf.Len() >= w.limit {
			return
		}
		remain := w.limit - w.buf.Len()
		if len(data) > remain {
			data = data[:remain]
		}
	}
	_, _ = w.buf.Write(data)
}

func (w *opsResponseCaptureWriter) bodyBytes() []byte {
	return w.buf.Bytes()
}

// OpsRequestTraceMiddleware captures a bounded preview of successful gateway responses
// so the Admin Ops request drill-down can show client-visible request/response details.
func OpsRequestTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c == nil || c.Writer == nil {
			c.Next()
			return
		}
		if _, ok := c.Writer.(*opsResponseCaptureWriter); ok {
			c.Next()
			return
		}
		captureWriter := newOpsResponseCaptureWriter(c.Writer, opsResponseCaptureLimit)
		c.Writer = captureWriter
		c.Set(opsResponseCaptureKey, captureWriter)
		c.Next()
	}
}

func buildCapturedUsagePayload(rawRequestBody []byte, c *gin.Context) (
	requestBody *string,
	requestBodyTruncated bool,
	requestBodyBytes *int,
	responseBody *string,
	responseBodyTruncated bool,
	responseBodyBytes *int,
) {
	requestBody, requestBodyTruncated, requestBodyBytes = service.PrepareAdminUsageRequestBody(rawRequestBody)
	if c == nil {
		return
	}
	raw, ok := c.Get(opsResponseCaptureKey)
	if !ok {
		return
	}
	captureWriter, ok := raw.(*opsResponseCaptureWriter)
	if !ok || captureWriter == nil {
		return
	}
	responseBody, responseBodyTruncated, responseBodyBytes = service.PrepareAdminUsageResponseBody(captureWriter.bodyBytes())
	return
}
