// Package lib provides OTel propagation helpers for Go backend.
package lib

import (
	"context"
	"os/exec"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// propagateContextToCmd injects the current OTel trace context from ctx
// into the environment of cmd as W3C TraceContext headers
// (TRACEPARENT / TRACESTATE).
//
// The spawned child process can then continue the trace if it uses an
// OTel SDK with a TraceContextPropagator.
//
// Usage:
//
//	cmd := exec.CommandContext(ctx, "git", "ls-remote", repoURL)
//	propagateContextToCmd(ctx, cmd)
func PropagateContextToCmd(ctx context.Context, cmd *exec.Cmd) {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	for key, value := range carrier {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
}
