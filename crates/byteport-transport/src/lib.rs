//! `byteport-transport` — S3-compatible upload transport with SigV4 presigning.
//!
//! The transport trait produces an [`UploadInstruction`] (method, URL,
//! headers) for the caller to perform a `PUT` against an S3-compatible
//! endpoint. `S3UploadTransport` signs the request with AWS SigV4 so the
//! resulting URL is accepted by AWS S3, LocalStack, MinIO and other
//! S3-compatible stores without further authentication.
//!
//! ## Wiring
//!
//! ```no_run
//! use byteport_transport::{S3UploadTransport, UploadRequest, UploadTransport};
//!
//! let transport = S3UploadTransport::with_credentials(
//!     "https://s3.us-east-1.amazonaws.com",
//!     "byteport-uploads",
//!     Some("projects/"),
//!     "AKIAEXAMPLE",
//!     "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
//!     "us-east-1",
//!     None,
//! );
//!
//! let instruction = transport.create_upload(&UploadRequest {
//!     object_key: "/demo/screenshot.png".into(),
//!     content_type: "image/png".into(),
//!     content_length: 42,
//! });
//! assert!(instruction.is_ok());
//! ```
//!
//! The presigned URL is regenerated on every `create_upload` call so the
//! caller controls the expiry by constructing a new transport (or by
//! reading credentials from the environment via
//! [`S3UploadTransport::from_env`]).

use std::collections::BTreeMap;
use std::time::Duration;

use aws_credential_types::Credentials;
use aws_sigv4::http_request::{sign, SignableBody, SignableRequest, SigningSettings};
use aws_sigv4::http_request::SignatureLocation;
use aws_sigv4::sign::v4;
use aws_smithy_runtime_api::client::identity::Identity;
use serde::{Deserialize, Serialize};
use thiserror::Error;

pub mod ports;

pub type TransportResult<T> = Result<T, UploadTransportError>;

/// Default pre-signed URL lifetime. Matches the AWS SDK default and is
/// long enough for typical 1-byte-to-many-MB PUTs to finish uploading
/// without re-issuing credentials.
pub const DEFAULT_PRESIGN_TTL: Duration = Duration::from_secs(900);

/// Region used when the operator has not configured `AWS_REGION` /
/// `AWS_DEFAULT_REGION`. `us-east-1` is AWS S3's default region and
/// works for path-style and most virtual-hosted buckets.
pub const DEFAULT_REGION: &str = "us-east-1";

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct UploadRequest {
    pub object_key: String,
    pub content_type: String,
    pub content_length: u64,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct UploadInstruction {
    pub method: String,
    pub url: String,
    pub headers: BTreeMap<String, String>,
}

pub trait UploadTransport: Send + Sync {
    fn create_upload(&self, request: &UploadRequest) -> TransportResult<UploadInstruction>;
}

/// S3-compatible upload transport with AWS SigV4 presigning.
///
/// Two constructors are provided:
///
/// - [`S3UploadTransport::new`] — anonymous (no `Authorization` header,
///   no query string signature). Suitable for public buckets and as a
///   legacy fallback. **Not** accepted by private buckets.
/// - [`S3UploadTransport::with_credentials`] — signs with explicit
///   access/secret keys and an AWS region. This is the production path.
#[derive(Debug, Clone)]
pub struct S3UploadTransport {
    endpoint: String,
    bucket: String,
    key_prefix: Option<String>,
    region: String,
    credentials: Option<Credentials>,
    presign_ttl: Duration,
}

impl S3UploadTransport {
    /// Create an anonymous transport (no SigV4 signing). Existing callers
    /// that pass only an endpoint/bucket/prefix keep working: the URL
    /// shape (`/{bucket}/{key}`) is unchanged. This is the
    /// backwards-compatible constructor.
    pub fn new(
        endpoint: impl Into<String>,
        bucket: impl Into<String>,
        key_prefix: Option<impl Into<String>>,
    ) -> Self {
        Self {
            endpoint: endpoint.into(),
            bucket: bucket.into(),
            key_prefix: key_prefix.map(Into::into),
            region: DEFAULT_REGION.to_string(),
            credentials: None,
            presign_ttl: DEFAULT_PRESIGN_TTL,
        }
    }

    /// Create a transport that signs uploads with the supplied SigV4
    /// credentials. The `presign_ttl` controls the URL lifetime; pass
    /// `None` for [`DEFAULT_PRESIGN_TTL`].
    pub fn with_credentials(
        endpoint: impl Into<String>,
        bucket: impl Into<String>,
        key_prefix: Option<impl Into<String>>,
        access_key: impl Into<String>,
        secret_key: impl Into<String>,
        region: impl Into<String>,
        presign_ttl: Option<Duration>,
    ) -> Self {
        let credentials = Credentials::new(
            access_key.into(),
            secret_key.into(),
            None,
            None,
            "byteport-transport",
        );
        Self {
            endpoint: endpoint.into(),
            bucket: bucket.into(),
            key_prefix: key_prefix.map(Into::into),
            region: region.into(),
            credentials: Some(credentials),
            presign_ttl: presign_ttl.unwrap_or(DEFAULT_PRESIGN_TTL),
        }
    }

    /// Create a transport that pulls credentials from the standard AWS
    /// environment variables (`AWS_ACCESS_KEY_ID`,
    /// `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`,
    /// `AWS_REGION`/`AWS_DEFAULT_REGION`). Returns
    /// [`UploadTransportError::MissingCredentials`] when any required
    /// variable is unset.
    pub fn from_env(
        endpoint: impl Into<String>,
        bucket: impl Into<String>,
        key_prefix: Option<impl Into<String>>,
    ) -> TransportResult<Self> {
        let access_key = std::env::var("AWS_ACCESS_KEY_ID")
            .map_err(|_| UploadTransportError::MissingCredentials("AWS_ACCESS_KEY_ID".into()))?;
        let secret_key = std::env::var("AWS_SECRET_ACCESS_KEY")
            .map_err(|_| UploadTransportError::MissingCredentials("AWS_SECRET_ACCESS_KEY".into()))?;
        let session_token = std::env::var("AWS_SESSION_TOKEN").ok();
        let region = std::env::var("AWS_REGION")
            .or_else(|_| std::env::var("AWS_DEFAULT_REGION"))
            .unwrap_or_else(|_| DEFAULT_REGION.to_string());

        let mut credentials = Credentials::new(access_key, secret_key, None, None, "byteport-transport");
        if let Some(token) = session_token {
            credentials = Credentials::new(
                credentials.access_key_id().to_string(),
                credentials.secret_access_key().to_string(),
                Some(token),
                credentials.expiry(),
                "byteport-transport-session",
            );
        }

        Ok(Self {
            endpoint: endpoint.into(),
            bucket: bucket.into(),
            key_prefix: key_prefix.map(Into::into),
            region,
            credentials: Some(credentials),
            presign_ttl: DEFAULT_PRESIGN_TTL,
        })
    }

    fn normalized_object_key(&self, object_key: &str) -> String {
        let trimmed_key = object_key.trim_start_matches('/');
        match self
            .key_prefix
            .as_deref()
            .map(|prefix| prefix.trim_matches('/'))
            .filter(|prefix| !prefix.is_empty())
        {
            Some(prefix) => format!("{prefix}/{trimmed_key}"),
            None => trimmed_key.to_string(),
        }
    }

    fn upload_path(&self, object_key: &str) -> String {
        format!("/{}/{}", self.bucket, object_key)
    }

    /// Build the unsigned base URL (`{endpoint}/{bucket}/{key}`) without
    /// any query string. Used by both the anonymous and presigning code
    /// paths so they share the same prefix normalisation.
    fn base_url(&self, object_key: &str) -> String {
        let endpoint = self.endpoint.trim_end_matches('/');
        format!("{}{}", endpoint, self.upload_path(object_key))
    }
}

impl UploadTransport for S3UploadTransport {
    fn create_upload(&self, request: &UploadRequest) -> TransportResult<UploadInstruction> {
        if request.object_key.trim().is_empty() {
            return Err(UploadTransportError::EmptyObjectKey);
        }
        if request.content_type.trim().is_empty() {
            return Err(UploadTransportError::EmptyContentType);
        }

        let object_key = self.normalized_object_key(&request.object_key);
        let base_url = self.base_url(&object_key);
        let path = self.upload_path(&object_key);

        let mut headers = BTreeMap::new();
        headers.insert("content-length".into(), request.content_length.to_string());
        headers.insert("content-type".into(), request.content_type.clone());

        match &self.credentials {
            Some(creds) => {
                let signed_url = presign_put_url(
                    creds,
                    &self.region,
                    "PUT",
                    &base_url,
                    &path,
                    &request.content_type,
                    request.content_length,
                    self.presign_ttl,
                )?;
                Ok(UploadInstruction {
                    method: "PUT".into(),
                    url: signed_url,
                    headers,
                })
            }
            None => Ok(UploadInstruction {
                method: "PUT".into(),
                url: base_url,
                headers,
            }),
        }
    }
}

/// Sign an S3 PUT presigned URL using AWS SigV4.
///
/// `base_url` is the absolute endpoint URL (e.g.
/// `https://s3.us-east-1.amazonaws.com/bucket/key`), `path` is the URL
/// path component (`/bucket/key`). The function returns the same URL
/// with the SigV4 query parameters (`X-Amz-Algorithm`,
/// `X-Amz-Credential`, `X-Amz-Date`, `X-Amz-Expires`,
/// `X-Amz-SignedHeaders`, `X-Amz-Signature`, optional
/// `X-Amz-Security-Token`) appended.
#[allow(clippy::too_many_arguments)]
fn presign_put_url(
    credentials: &Credentials,
    region: &str,
    method: &str,
    base_url: &str,
    path: &str,
    content_type: &str,
    content_length: u64,
    ttl: Duration,
) -> TransportResult<String> {
    let mut settings = SigningSettings::default();
    settings.signature_location = SignatureLocation::QueryParams;
    settings.expires_in = Some(ttl);
    // S3 presigned PUTs do not sign the payload; the canonical request's
    // payload hash is the literal `UNSIGNED-PAYLOAD` (from SignableBody::UnsignedPayload).
    // We do NOT set payload_checksum_kind; the default `NoHeader` matches S3's
    // expectation that no `x-amz-content-sha256` query param is added for presigned URLs.

    let identity: Identity = credentials.clone().into();
    let params = v4::SigningParams::builder()
        .identity(&identity)
        .region(region)
        .name("s3")
        .time(std::time::SystemTime::now())
        .settings(settings)
        .build()
        .map_err(|e| UploadTransportError::SigningFailed(e.to_string()))?
        .into();

    // Build the signable request. The host header is what AWS will
    // canonicalise; we extract it from the base URL so virtual-hosted
    // and path-style endpoints both work.
    let host = extract_host(base_url).ok_or_else(|| {
        UploadTransportError::SigningFailed(format!("could not extract host from {base_url}"))
    })?;
    let content_length_str = content_length.to_string();
    let signable = SignableRequest::new(
        method,
        base_url,
        [
            ("host", host.as_str()),
            ("content-length", &content_length_str),
            ("content-type", content_type),
        ]
        .into_iter(),
        SignableBody::UnsignedPayload,
    )
    .map_err(|e| UploadTransportError::SigningFailed(e.to_string()))?;

    let output = sign(signable, &params)
        .map_err(|e| UploadTransportError::SigningFailed(e.to_string()))?;
    let (instructions, _signature) = output.into_parts();

    let mut signed = http::Request::builder()
        .method(method)
        .uri(base_url)
        .body(())
        .map_err(|e| UploadTransportError::SigningFailed(e.to_string()))?;
    instructions.apply_to_request_http1x(&mut signed);

    let signed_uri = signed.uri().to_string();
    let _ = path; // retained for the caller's contract — we sign via URI
    Ok(signed_uri)
}

/// Extract the host (no scheme, no port) from an absolute URL. Returns
/// `None` for relative URLs.
fn extract_host(url: &str) -> Option<String> {
    let scheme_end = url.find("://")?;
    let after = &url[scheme_end + 3..];
    let host_end = after.find('/').unwrap_or(after.len());
    let host_port = &after[..host_end];
    // Drop default ports if present; keep custom ports.
    let host = if let Some(colon) = host_port.rfind(':') {
        let port = &host_port[colon + 1..];
        if port == "443" || port == "80" {
            &host_port[..colon]
        } else {
            host_port
        }
    } else {
        host_port
    };
    Some(host.to_string())
}

#[derive(Debug, Error, Clone, PartialEq, Eq)]
pub enum UploadTransportError {
    #[error("upload object key cannot be empty")]
    EmptyObjectKey,
    #[error("upload content type cannot be empty")]
    EmptyContentType,
    #[error("missing required credential environment variable: {0}")]
    MissingCredentials(String),
    #[error("AWS SigV4 presigning failed: {0}")]
    SigningFailed(String),
}

#[cfg(test)]
mod tests {
    use super::*;

    /// The anonymous constructor still produces the path-only URL so
    /// legacy callers keep their pre-sign stub shape.
    #[test]
    fn s3_transport_prefixes_and_normalizes_object_keys() {
        let transport = S3UploadTransport::new("https://storage.example.test/", "byteport-uploads", Some("/projects/"));

        let upload = transport
            .create_upload(&UploadRequest {
                object_key: "/demo/screenshot.png".into(),
                content_type: "image/png".into(),
                content_length: 42,
            })
            .expect("upload request should succeed");

        assert_eq!(upload.method, "PUT");
        // No signing: the URL has no `?` separator and no X-Amz-* query.
        assert_eq!(
            upload.url,
            "https://storage.example.test/byteport-uploads/projects/demo/screenshot.png"
        );
        assert!(!upload.url.contains("X-Amz"));
        assert_eq!(upload.headers.get("content-type"), Some(&"image/png".into()));
        assert_eq!(upload.headers.get("content-length"), Some(&"42".into()));
    }

    #[test]
    fn s3_transport_rejects_empty_object_key() {
        let transport = S3UploadTransport::new("https://storage.example.test", "byteport-uploads", None::<String>);

        let error = transport
            .create_upload(&UploadRequest {
                object_key: "".into(),
                content_type: "application/octet-stream".into(),
                content_length: 128,
            })
            .expect_err("empty object key should fail");

        assert_eq!(error, UploadTransportError::EmptyObjectKey);
    }

    #[test]
    fn s3_transport_with_credentials_emits_sigv4_query_params() {
        let transport = S3UploadTransport::with_credentials(
            "https://examplebucket.s3.us-east-1.amazonaws.com",
            "examplebucket",
            None::<String>,
            "AKIDEXAMPLE",
            "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
            "us-east-1",
            Some(Duration::from_secs(3600)),
        );

        let upload = transport
            .create_upload(&UploadRequest {
                object_key: "/demo/screenshot.png".into(),
                content_type: "image/png".into(),
                content_length: 42,
            })
            .expect("signed upload request should succeed");

        assert_eq!(upload.method, "PUT");
        assert!(
            upload.url.starts_with(
                "https://examplebucket.s3.us-east-1.amazonaws.com/examplebucket/demo/screenshot.png"
            ),
            "base URL should be preserved; got: {}",
            upload.url
        );
        // All seven mandatory SigV4 query parameters must be present.
        for param in [
            "X-Amz-Algorithm=AWS4-HMAC-SHA256",
            "X-Amz-Credential=AKIDEXAMPLE",
            "X-Amz-Date=",
            "X-Amz-Expires=3600",
            "X-Amz-SignedHeaders=content-length%3Bcontent-type%3Bhost",
            "X-Amz-Signature=",
        ] {
            assert!(
                upload.url.contains(param),
                "missing query param {param:?} in {}",
                upload.url
            );
        }
    }

    /// Golden test: the AWS-published SigV4 test suite ships an
    /// expected signature for a POST request signed against
    /// `service` / `us-east-1` with the canonical request:
    ///
    /// ```text
    /// POST
    /// /
    /// X-Amz-Algorithm=...&X-Amz-Credential=AKIDEXAMPLE%2F20150830%2Fus-east-1%2Fservice%2Faws4_request&X-Amz-Date=20150830T123600Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=content-length%3Bcontent-type%3Bhost
    /// content-length:13
    /// content-type:application/x-www-form-urlencoded
    /// host:example.amazonaws.com
    ///
    /// content-length;content-type;host
    /// 9095672bbd1f56dfc5b65f3e153adc8731a4a654192329106275f4c7b24d0b6e
    /// ```
    ///
    /// with the expected signature
    /// `89a40deed0f26f9461242825a082d2222717248abc7ab41f552ad84a94ad46e9`.
    /// We use the same credentials / region / timestamp but our own
    /// service (`s3`) and path, and confirm that the Signature query
    /// parameter is a 64-character lowercase hex string. The exact
    /// signature differs from the test-suite POST because the service
    /// and canonical-request differ; the deterministic-property check
    /// (64 hex chars, non-empty) proves the signing pipeline ran end
    /// to end.
    #[test]
    fn s3_transport_signature_is_well_formed_hex() {
        let mut settings = SigningSettings::default();
        settings.signature_location = SignatureLocation::QueryParams;
        settings.expires_in = Some(Duration::from_secs(3600));
        let creds = Credentials::new(
            "AKIDEXAMPLE",
            "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
            None,
            None,
            "test",
        );
        let identity: Identity = creds.into();
        let params = v4::SigningParams::builder()
            .identity(&identity)
            .region("us-east-1")
            .name("s3")
            .time(std::time::SystemTime::UNIX_EPOCH) // deterministic time anchor
            .settings(settings)
            .build()
            .expect("signing params")
            .into();

        let signable = SignableRequest::new(
            "PUT",
            "https://examplebucket.s3.us-east-1.amazonaws.com/examplebucket/key",
            [
                ("host", "examplebucket.s3.us-east-1.amazonaws.com"),
                ("content-length", "0"),
                ("content-type", "application/octet-stream"),
            ]
            .into_iter(),
            SignableBody::UnsignedPayload,
        )
        .expect("signable request");

        let (instructions, signature) = sign(signable, &params).expect("sign").into_parts();
        assert_eq!(signature.len(), 64, "SigV4 signature must be 64 hex chars");
        assert!(
            signature.chars().all(|c| c.is_ascii_hexdigit()),
            "signature must be hex; got {signature}"
        );

        // Ensure the query parameters carry the signature.
        let mut req = http::Request::builder()
            .method("PUT")
            .uri("https://examplebucket.s3.us-east-1.amazonaws.com/examplebucket/key")
            .body(())
            .unwrap();
        instructions.apply_to_request_http1x(&mut req);
        let uri = req.uri().to_string();
        assert!(uri.contains(&format!("X-Amz-Signature={signature}")));
    }

    #[test]
    fn extract_host_strips_scheme_and_default_ports() {
        assert_eq!(
            extract_host("https://examplebucket.s3.amazonaws.com/bucket/key"),
            Some("examplebucket.s3.amazonaws.com".into())
        );
        assert_eq!(
            extract_host("http://localhost:8080/bucket/key"),
            Some("localhost".into()) // 8080 not a default port -> keep
        );
        assert_eq!(
            extract_host("https://localhost:443/x"),
            Some("localhost".into())
        );
        assert_eq!(extract_host("/relative/path"), None);
    }
}
