package tracing

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type httpHandler struct {
	name    string
	handler http.Handler
}

func NewHttpHandler(name string, h http.Handler) http.Handler {
	return &httpHandler{name: name, handler: h}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx, span := otel.Tracer(h.name).Start(ctx, h.name)
	defer span.End()

	h.handler.ServeHTTP(w, req.WithContext(ctx))
	log.Infof(fmt.Sprintf("[trace_id=%s]", span.SpanContext().TraceID().String()))
}
