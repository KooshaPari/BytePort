//! Criterion benchmarks for `byteport-dag`.
//!
//! Measures latency of DAG node creation, topological sort, and traversal.

use criterion::{black_box, criterion_group, criterion_main, Criterion};

use byteport_dag::dag::Dag;
use byteport_dag::topo::{dfs_sort, kahn_sort};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Build a chain DAG with `n` nodes: 0 → 1 → 2 → … → n-1.
fn chain_dag(n: usize) -> Dag<usize> {
    let mut dag = Dag::new();
    for i in 0..n {
        dag.add_node(i).unwrap();
    }
    for i in 0..n.saturating_sub(1) {
        dag.add_edge(i, i + 1).unwrap();
    }
    dag
}

/// Build a "wide" DAG with `n` nodes where the first is a root fanning out
/// to all remaining leaves.  0 → 1, 0 → 2, …, 0 → n-1.
fn star_dag(n: usize) -> Dag<usize> {
    let mut dag = Dag::new();
    for i in 0..n {
        dag.add_node(i).unwrap();
    }
    for i in 1..n {
        dag.add_edge(0, i).unwrap();
    }
    dag
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

/// Build a DAG by inserting N nodes sequentially.
fn bench_node_creation(c: &mut Criterion) {
    let mut group = c.benchmark_group("dag/node_creation");

    for &n in &[100usize, 1000] {
        group.bench_function(format!("{n}_nodes"), |b| {
            b.iter(|| {
                let mut dag = Dag::new();
                for i in 0..black_box(n) {
                    dag.add_node(black_box(i)).unwrap();
                }
            });
        });
    }
    group.finish();
}

/// Build a DAG with N nodes and N-1 edges (chain topology).
fn bench_edge_creation(c: &mut Criterion) {
    let mut group = c.benchmark_group("dag/edge_creation");

    for &n in &[100usize, 1000] {
        group.bench_function(format!("{n}_nodes"), |b| {
            b.iter(|| {
                let mut dag = Dag::new();
                for i in 0..n {
                    dag.add_node(black_box(i)).unwrap();
                }
                for i in 0..n.saturating_sub(1) {
                    dag.add_edge(black_box(i), black_box(i + 1)).unwrap();
                }
            });
        });
    }
    group.finish();
}

/// Topological sort of a chain DAG using Kahn's algorithm.
fn bench_kahn_sort_chain(c: &mut Criterion) {
    let dag = chain_dag(1000);

    c.bench_function("dag/kahn_sort_chain_1000", |b| {
        b.iter(|| {
            let order = kahn_sort(black_box(&dag)).unwrap();
            black_box(order.len());
        });
    });
}

/// Topological sort of a star DAG using Kahn's algorithm.
fn bench_kahn_sort_star(c: &mut Criterion) {
    let dag = star_dag(1000);

    c.bench_function("dag/kahn_sort_star_1000", |b| {
        b.iter(|| {
            let order = kahn_sort(black_box(&dag)).unwrap();
            black_box(order.len());
        });
    });
}

/// Topological sort of a chain DAG using DFS-based sort.
fn bench_dfs_sort_chain(c: &mut Criterion) {
    let dag = chain_dag(1000);

    c.bench_function("dag/dfs_sort_chain_1000", |b| {
        b.iter(|| {
            let order = dfs_sort(black_box(&dag)).unwrap();
            black_box(order.len());
        });
    });
}

/// Topological sort of a star DAG using DFS-based sort.
fn bench_dfs_sort_star(c: &mut Criterion) {
    let dag = star_dag(1000);

    c.bench_function("dag/dfs_sort_star_1000", |b| {
        b.iter(|| {
            let order = dfs_sort(black_box(&dag)).unwrap();
            black_box(order.len());
        });
    });
}

/// Traverse all nodes via `iter_nodes` and all edges via `children_of`.
fn bench_traversal(c: &mut Criterion) {
    let dag = chain_dag(1000);

    c.bench_function("dag/traversal_1000", |b| {
        b.iter(|| {
            for node in dag.iter_nodes() {
                black_box(node);
                if let Some(children) = dag.children_of(node) {
                    for child in children {
                        black_box(child);
                    }
                }
            }
        });
    });
}

// ---------------------------------------------------------------------------
// Criterion registration
// ---------------------------------------------------------------------------

criterion_group!(
    benches,
    bench_node_creation,
    bench_edge_creation,
    bench_kahn_sort_chain,
    bench_kahn_sort_star,
    bench_dfs_sort_chain,
    bench_dfs_sort_star,
    bench_traversal,
);
criterion_main!(benches);
