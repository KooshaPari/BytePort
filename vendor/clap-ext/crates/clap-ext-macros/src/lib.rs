//! Procedural macros for clap-ext.

use proc_macro::TokenStream;
use quote::quote;

/// Marker attribute: tag a `clap::Subcommand` enum to declare it
/// accepts the 3 common subcommands (`init`, `validate`, `version`).
///
/// Currently a passthrough — the user is expected to add the variants
/// manually and may use this attribute for documentation/intent.
///
/// Usage:
/// ```ignore
/// #[derive(Subcommand)]
/// #[clap_ext_common_subcommands]
/// enum Commands {
///     Init(InitCmd),
///     Validate(ValidateCmd),
///     Version(VersionCmd),
///     // ... your other variants
/// }
/// ```
#[proc_macro_attribute]
pub fn clap_ext_common_subcommands(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let output: proc_macro2::TokenStream = item.into();
    TokenStream::from(quote! {
        #[doc = "Enum carries the 3 clap-ext common subcommands (init/validate/version)."]
        #output
    })
}
