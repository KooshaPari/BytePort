use std::time::Duration;

use aws_credential_types::Credentials;
use aws_sigv4::http_request::{sign, SignableBody, SignableRequest, SignatureLocation, SigningSettings};
use aws_sigv4::sign::v4;
use aws_smithy_runtime_api::client::identity::Identity;

use crate::{TransportResult, UploadTransportError};

/// Sign an S3 PUT URL with AWS SigV4 query parameters.
#[allow(clippy::too_many_arguments)]
pub(super) fn presign_put_url(
    credentials: &Credentials,
    region: &str,
    method: &str,
    base_url: &str,
    path: &str,
    content_type: &str,
    content_length: u64,
    ttl: Duration,
) -> TransportResult<String> {
    validate_region(region)?;
    let signing_time = std::time::SystemTime::now();
    if credentials.expiry().is_some_and(|expiry| expiry <= signing_time) {
        return Err(UploadTransportError::SigningFailed(
            "AWS credentials have expired".into(),
        ));
    }

    let mut settings = SigningSettings::default();
    settings.signature_location = SignatureLocation::QueryParams;
    settings.expires_in = Some(ttl);
    // Presigned S3 PUT requests use an unsigned payload and omit the checksum header.
    let identity: Identity = credentials.clone().into();
    let params = v4::SigningParams::builder()
        .identity(&identity)
        .region(region)
        .name("s3")
        .time(signing_time)
        .settings(settings)
        .build()
        .map_err(|error| UploadTransportError::SigningFailed(error.to_string()))?
        .into();

    let host = extract_host(base_url)
        .ok_or_else(|| UploadTransportError::SigningFailed(format!("could not extract host from {base_url}")))?;
    let content_length = content_length.to_string();
    let signable = SignableRequest::new(
        method,
        base_url,
        [
            ("host", host.as_str()),
            ("content-length", content_length.as_str()),
            ("content-type", content_type),
        ]
        .into_iter(),
        SignableBody::UnsignedPayload,
    )
    .map_err(|error| UploadTransportError::SigningFailed(error.to_string()))?;

    let output = sign(signable, &params).map_err(|error| UploadTransportError::SigningFailed(error.to_string()))?;
    let (instructions, signature) = output.into_parts();
    if signature.is_empty() {
        return Err(UploadTransportError::SigningFailed(
            "AWS SigV4 presigning produced an empty signature".into(),
        ));
    }

    let mut signed = http::Request::builder()
        .method(method)
        .uri(base_url)
        .body(())
        .map_err(|error| UploadTransportError::SigningFailed(error.to_string()))?;
    instructions.apply_to_request_http1x(&mut signed);

    let signed_url = signed.uri().to_string();
    ensure_signature_present(&signed_url)?;
    let _ = path; // The URI carries the path used by the signer.
    Ok(signed_url)
}

fn validate_region(region: &str) -> TransportResult<()> {
    let valid = !region.is_empty()
        && region
            .bytes()
            .all(|byte| byte.is_ascii_lowercase() || byte.is_ascii_digit() || byte == b'-');
    if valid {
        Ok(())
    } else {
        Err(UploadTransportError::SigningFailed(format!(
            "invalid AWS region: {region:?}"
        )))
    }
}

fn ensure_signature_present(url: &str) -> TransportResult<()> {
    let has_signature = url
        .split_once('?')
        .map(|(_, query)| {
            query.split('&').any(|parameter| {
                parameter
                    .split_once('=')
                    .is_some_and(|(name, value)| name == "X-Amz-Signature" && !value.is_empty())
            })
        })
        .unwrap_or(false);
    if has_signature {
        Ok(())
    } else {
        Err(UploadTransportError::SigningFailed(
            "AWS SigV4 presigning produced no signature query parameter".into(),
        ))
    }
}

fn extract_host(url: &str) -> Option<String> {
    let scheme_end = url.find("://")?;
    let after = &url[scheme_end + 3..];
    let host_end = after.find('/').unwrap_or(after.len());
    let host_port = &after[..host_end];
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

#[cfg(test)]
mod tests {
    use super::*;

    fn credentials(expiry: Option<std::time::SystemTime>) -> Credentials {
        Credentials::new(
            "AKIDEXAMPLE",
            "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
            None,
            expiry,
            "presign-unit-test",
        )
    }

    fn presign(credentials: &Credentials, region: &str, endpoint: &str) -> TransportResult<String> {
        let url = format!("{endpoint}/byteport-uploads/test/object.bin");
        presign_put_url(
            credentials,
            region,
            "PUT",
            &url,
            "/byteport-uploads/test/object.bin",
            "application/octet-stream",
            17,
            Duration::from_secs(900),
        )
    }

    #[test]
    fn presign_put_url_success_returns_signed_put_url() {
        let url = presign(&credentials(None), "us-east-1", "https://s3.us-east-1.amazonaws.com")
            .expect("presigning should succeed");

        assert!(url.starts_with("https://s3.us-east-1.amazonaws.com/byteport-uploads/test/object.bin"));
        assert!(url.contains("X-Amz-Algorithm=AWS4-HMAC-SHA256"));
        assert!(url.contains("X-Amz-Expires=900"));
        assert!(ensure_signature_present(&url).is_ok());
    }

    #[test]
    fn presign_rejects_missing_signature() {
        let error =
            ensure_signature_present("https://s3.example.test/bucket/key?X-Amz-Date=20261018T000000Z&X-Amz-Signature=")
                .expect_err("empty signature must be rejected");

        assert!(error.to_string().contains("no signature query parameter"));
    }

    #[test]
    fn presign_rejects_invalid_region() {
        let error = presign(&credentials(None), "not a region", "https://s3.example.test")
            .expect_err("invalid region must be rejected");

        assert!(error.to_string().contains("invalid AWS region"));
    }

    #[test]
    fn presign_rejects_expired_credentials() {
        let error = presign(
            &credentials(Some(std::time::SystemTime::UNIX_EPOCH)),
            "us-east-1",
            "https://s3.example.test",
        )
        .expect_err("expired credentials must be rejected");

        assert!(error.to_string().contains("credentials have expired"));
    }

    #[test]
    fn presign_supports_localstack_endpoint_with_custom_port() {
        let url = presign(&credentials(None), "us-east-1", "http://localhost:4566")
            .expect("LocalStack endpoint should be presignable");

        assert!(url.starts_with("http://localhost:4566/byteport-uploads/test/object.bin"));
        assert_eq!(
            extract_host("http://localhost:4566/byteport-uploads/test/object.bin"),
            Some("localhost:4566".into())
        );
        assert!(ensure_signature_present(&url).is_ok());
    }

    #[test]
    fn s3_transport_signature_is_well_formed_hex() {
        let mut settings = SigningSettings::default();
        settings.signature_location = SignatureLocation::QueryParams;
        settings.expires_in = Some(Duration::from_secs(3600));
        let identity: Identity = credentials(None).into();
        let params = v4::SigningParams::builder()
            .identity(&identity)
            .region("us-east-1")
            .name("s3")
            .time(std::time::SystemTime::UNIX_EPOCH)
            .settings(settings)
            .build()
            .expect("signing params")
            .into();
        let url = "https://examplebucket.s3.us-east-1.amazonaws.com/examplebucket/key";
        let signable = SignableRequest::new(
            "PUT",
            url,
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
        assert_eq!(signature.len(), 64);
        assert!(signature.chars().all(|character| character.is_ascii_hexdigit()));

        let mut request = http::Request::builder()
            .method("PUT")
            .uri(url)
            .body(())
            .expect("request");
        instructions.apply_to_request_http1x(&mut request);
        assert!(request
            .uri()
            .to_string()
            .contains(&format!("X-Amz-Signature={signature}")));
    }

    #[test]
    fn extract_host_strips_scheme_and_default_ports() {
        assert_eq!(
            extract_host("https://examplebucket.s3.amazonaws.com/bucket/key"),
            Some("examplebucket.s3.amazonaws.com".into())
        );
        assert_eq!(
            extract_host("http://localhost:8080/bucket/key"),
            Some("localhost:8080".into())
        );
        assert_eq!(extract_host("https://localhost:443/x"), Some("localhost".into()));
        assert_eq!(extract_host("/relative/path"), None);
    }
}
