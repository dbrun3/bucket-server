# Bucket Server

S3-backed file server with caching and SPA support. Host multiple sites from private S3 buckets, avoiding public S3 data transfer costs and accidental exposure risks. Includes optional built-in caching with a stale refresh period, and the option to additionally cache content in the browser.

## Bucket Naming

The `Host` header determines the S3 bucket name:
- `example.com` → bucket `example.com`
- `docs.example.com` → bucket `docs.example.com`

## Recommended Setup

**Recommended Configuration using an internal ALB:**
- Wildcard DNS: `*.example.com` → ALB
- Wildcard SSL certificate for `*.example.com`
- Forward all traffic to bucket server on port 8080
- Health check: `/health`

The ALB preserves the `Host` header, allowing the server to route to the correct S3 bucket. Add new subdomains by creating matching S3 buckets—no ALB changes needed.