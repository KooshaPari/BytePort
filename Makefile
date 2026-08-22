.PHONY: build test lint clean

# BytePort multi-language platform

build:
	cargo build --release

test:
	cargo test

lint:
	cargo clippy -- -D warnings
	cargo fmt --check

clean:
	cargo clean
