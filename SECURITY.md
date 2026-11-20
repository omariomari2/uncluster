# Security policy

## Reporting a vulnerability

Please report security issues privately through GitHub's private vulnerability reporting:

https://github.com/omariomari2/uncluster/security/advisories/new

Do not open a public issue, pull request, or discussion for a vulnerability.

Include enough detail to reproduce the problem: the affected command or API route, input, and what you expected versus what happened. You should receive an acknowledgment after the report is reviewed.

## Remote request controls

Packages that fetch remote pages or assets (`internal/fetcher` and `internal/scraper`) go through `internal/safehttp`:

- Only `http` and `https` URLs with a host are accepted.
- Loopback, private, link-local, multicast, unspecified, and reserved destinations are blocked at connection time.
- Redirects are revalidated, and at most 10 redirects are followed.
- Each response body is capped at 25 MiB.

These controls reduce the risk of server-side request forgery and oversized downloads. They do not grant permission to copy a website, and they are not a complete security audit of the project.
