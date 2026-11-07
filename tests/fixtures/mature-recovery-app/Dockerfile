FROM nginx:alpine

ARG SOURCE_COMMIT=UNSET
ARG MANIFEST_DIGEST=UNSET
ARG FIXTURE_NONCE=UNSET

RUN printf '%s\n' \
  '{' \
  '  "fixture": "byteport-mature-recovery",' \
  '  "source_commit": "'"$SOURCE_COMMIT"'",' \
  '  "manifest_digest": "'"$MANIFEST_DIGEST"'",' \
  '  "fixture_nonce": "'"$FIXTURE_NONCE"'"' \
  '}' > /usr/share/nginx/html/__byteport_probe

EXPOSE 80
