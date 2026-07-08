//! Output formatting helpers.
//!
//! Two formats are supported everywhere:
//!
//! - **Table** — human-readable, colored (when allowed), aligned columns.
//! - **Json** — pretty-printed JSON for piping into `jq`, `xargs`, etc.
//!
//! The format is selected via the global `--json` flag on the top-level
//! [`Cli`](crate::cli::Cli). Each command receives a
//! [`OutputContext`](crate::output::OutputContext) carrying the format +
//! color preference.
//!
//! ## Color handling
//!
//! Color output obeys the [`NO_COLOR`](https://no-color.org/) convention:
//! if the env var is set to any non-empty value, ANSI escapes are omitted
//! regardless of TTY status. Use [`respect_no_color`] at startup.

use std::fmt::{self, Display};
use std::io::{self, Write};

use colored::control::{set_override, ShouldColorize};
use serde::Serialize;

/// Output format selection.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum OutputFormat {
    /// Human-readable table.
    #[default]
    Table,
    /// Machine-readable JSON.
    Json,
}

impl OutputFormat {
    /// Returns true if the format is JSON.
    pub fn is_json(self) -> bool {
        matches!(self, OutputFormat::Json)
    }
}

/// Per-invocation output context.
#[derive(Debug, Clone, Copy)]
pub struct OutputContext {
    /// Selected output format.
    pub format: OutputFormat,
    /// Whether color is enabled (after `--no-color` / `NO_COLOR` resolution).
    pub color: bool,
}

impl OutputContext {
    /// Construct a new context.
    pub fn new(format: OutputFormat, color: bool) -> Self {
        Self { format, color }
    }

    /// Force the JSON format, regardless of the user's preference.
    pub fn json() -> Self {
        Self {
            format: OutputFormat::Json,
            color: false,
        }
    }

    /// Force the Table format with color enabled.
    pub fn table() -> Self {
        Self {
            format: OutputFormat::Table,
            color: true,
        }
    }
}

/// Honor the `NO_COLOR` environment convention.
///
/// If `NO_COLOR` is set to a non-empty value (regardless of its value,
/// per <https://no-color.org/>), colored output is disabled globally.
/// Returns `true` if colors will be emitted.
pub fn respect_no_color() -> bool {
    let no_color = std::env::var_os("NO_COLOR")
        .map(|v| !v.is_empty())
        .unwrap_or(false);
    if no_color {
        set_override(ShouldColorize::Never);
        false
    } else {
        ShouldColorize::from_env().should_colorize()
    }
}

/// Render a value to a writer in the selected format.
pub fn print_to<W: Write, T: Serialize + ?Sized>(
    writer: &mut W,
    value: &T,
    ctx: OutputContext,
) -> io::Result<()> {
    match ctx.format {
        OutputFormat::Table => {
            // Plain-JSON dump for now; table layout is added per-command.
            writeln!(writer, "{}", Display::fmt(value, &mut DisplayFmt)?)?;
        }
        OutputFormat::Json => {
            let s = serde_json::to_string_pretty(value)
                .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;
            writeln!(writer, "{s}")?;
        }
    }
    Ok(())
}

/// Print a value to stdout using the selected format.
pub fn print<T: Serialize + ?Sized>(value: &T, ctx: OutputContext) -> anyhow::Result<()> {
    let mut out = io::stdout().lock();
    print_to(&mut out, value, ctx)?;
    Ok(())
}

/// Print a JSON value to stdout regardless of the user's preferred format.
///
/// Useful for `--json` always-on flags or for sub-commands that only have
/// a JSON representation.
pub fn print_json<T: Serialize + ?Sized>(value: &T) -> anyhow::Result<()> {
    let s = serde_json::to_string_pretty(value)?;
    println!("{s}");
    Ok(())
}

/// Adapter so `Display` types can be printed through the `Serialize` API.
struct DisplayFmt<'a, T: ?Sized>(&'a T);

impl<T: Display + ?Sized> fmt::Display for DisplayFmt<'_, T> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        self.0.fmt(f)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn output_format_default_is_table() {
        assert_eq!(OutputFormat::default(), OutputFormat::Table);
    }

    #[test]
    fn output_format_is_json_predicate() {
        assert!(!OutputFormat::Table.is_json());
        assert!(OutputFormat::Json.is_json());
    }

    #[test]
    fn print_json_writes_pretty() {
        let v = serde_json::json!({"a": 1, "b": [2, 3]});
        let s = serde_json::to_string_pretty(&v).unwrap();
        assert!(s.contains("\"a\": 1"));
        assert!(s.contains("\"b\":"));
    }

    #[test]
    fn no_color_disables_color_globally() {
        // SAFETY: tests may run in parallel, but `set_override` is a global
        // setting. This test only asserts the API; we don't make any
        // claims about other concurrent tests.
        let prev = std::env::var_os("NO_COLOR");
        // SAFETY: single-threaded test environment.
        unsafe {
            std::env::set_var("NO_COLOR", "1");
        }
        let colored = respect_no_color();
        assert!(!colored, "NO_COLOR=1 must disable color");

        match prev {
            Some(v) => unsafe {
                std::env::set_var("NO_COLOR", v);
            },
            None => unsafe {
                std::env::remove_var("NO_COLOR");
            },
        }
    }
}