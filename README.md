# DMARC Verifier (simple)

A tiny REST server that verifies a domain's DMARC record from an email.

## Setup

Prerequisites: Go installed (1.18+).

Run locally:

```bash
# PowerShell:
.\webserver.exe

# web server will be active on http://localhost:8080
```

## Api Endpoint

POST /verify

Request JSON:

```json
{ "email": "user@example.com" }
```

Response JSON examples:

- DMARC found

```json
{
  "dmarcRecord": "v=DMARC1; p=reject; rua=mailto:example@example.com",
  "isDmarc": true,
  "dmarcType": "reject"
}
```

- DMARC not found

```json
{
  "dmarcRecord": "",
  "isDmarc": false,
  "dmarcType": "none"
}
```

### Detailed documentation

if you are interested to understand the project more deeply, do check out the technical documentation on the deepwiki page of our project [here](https://deepwiki.com/Gitax18/email-verifier)
