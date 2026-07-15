//! Criterion benchmarks for `byteport-otel`.
//!
//! Measures latency of span creation, context propagation, attribute
//! attachment, and concurrent span creation.

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use opentelemetry::trace::{Span, SpanKind, TraceContextExt, Tracer, TracerProvider as _};
use opentelemetry::Context;
use opentelemetry_sdk::trace::SdkTracerProvider;

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn build_provider() -> SdkTracerProvider {
    SdkTracerProvider::builder().build()
}

fn make_tracer(provider: &SdkTracerProvider) -> impl Tracer + '_ {
    provider.tracer("bench")
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

/// Basic span creation and immediate closure (no attributes, no children).
fn bench_span_create_and_close(c: &mut Criterion) {
    let provider = build_provider();
    let tracer = make_tracer(&provider);

    c.bench_function("otel/span_create_and_close", |b| {
        b.iter(|| {
            let mut span = tracer.start(black_box("bench-span"));
            span.end();
        });
    });
}

/// Create a parent span, then a child span attached to the parent's context.
fn bench_context_propagation(c: &mut Criterion) {
    let provider = build_provider();
    let tracer = make_tracer(&provider);

    c.bench_function("otel/context_propagation", |b| {
        b.iter(|| {
            let parent = tracer.start(black_box("parent"));
            let cx = Context::current().with_span(parent);
            let guard = cx.attach();

            let mut child = tracer
                .span_builder(black_box("child"))
                .with_kind(SpanKind::Internal)
                .start_with_context(&tracer, &cx);

            child.end();
            drop(guard);
        });
    });
}

/// Create a span and attach N key-value attributes.
fn bench_attributes(c: &mut Criterion) {
    let provider = build_provider();
    let tracer = make_tracer(&provider);

    let mut group = c.benchmark_group("otel/attributes");
    for &count in &[5, 10, 50] {
        group.bench_function(format!("{count}_attrs"), |b| {
            b.iter(|| {
                let mut span = tracer.start(black_box("attr-span"));
                for i in 0..count {
                    span.set_attribute(opentelemetry::KeyValue::new(
                        black_box(format!("key_{i}")),
                        black_box(format!("val_{i}")),
                    ));
                }
                span.end();
            });
        });
    }
    group.finish();
}

/// Spawn 10 concurrent tasks, each creating and closing a span.
///
/// Uses `std::thread::scope` so the main thread waits for all workers
/// before criterion measures the wall-clock time of parallel span creation.
fn bench_concurrent_spans(c: &mut Criterion) {
    let provider = build_provider();

    c.bench_function("otel/concurrent_spans_10", |b| {
        b.iter(|| {
            std::thread::scope(|s| {
                for i in 0..10 {
                    // Each thread gets its own tracer from the shared provider.
                    // OTel SDK tracers are Send + Clone so this is safe.
                    let tracer = provider.tracer("bench");
                    s.spawn(move || {
                        let mut span = tracer.start(black_box(format!("conc-t{i}")));
                        span.end();
                    });
                }
            });
        });
    });
}

// ---------------------------------------------------------------------------
// Criterion registration
// ---------------------------------------------------------------------------

criterion_group!(
    benches,
    bench_span_create_and_close,
    bench_context_propagation,
    bench_attributes,
    bench_concurrent_spans,
);
criterion_main!(benches);
