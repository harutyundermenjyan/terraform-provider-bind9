---
page_title: "bind9_record Resource - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Manages DNS records in a zone on BIND9 server.
---

# bind9_record (Resource)

Manages DNS records in a zone on a BIND9 server. Supports all common DNS record types including A, AAAA, CNAME, MX, TXT, SRV, CAA, and more.

## Example Usage

### A Record (IPv4)

```hcl
resource "bind9_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["192.168.1.100"]
}

# Multiple A records (round-robin)
resource "bind9_record" "app" {
  zone    = "example.com"
  name    = "app"
  type    = "A"
  ttl     = 60
  records = [
    "192.168.1.101",
    "192.168.1.102",
    "192.168.1.103",
  ]
}
```

### AAAA Record (IPv6)

```hcl
resource "bind9_record" "www_ipv6" {
  zone    = "example.com"
  name    = "www"
  type    = "AAAA"
  ttl     = 300
  records = ["2001:db8::100"]
}
```

### CNAME Record (Alias)

```hcl
resource "bind9_record" "blog" {
  zone    = "example.com"
  name    = "blog"
  type    = "CNAME"
  ttl     = 3600
  records = ["www.example.com."]
}
```

### MX Record (Mail)

```hcl
resource "bind9_record" "mx" {
  zone    = "example.com"
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = [
    "10 mail1.example.com.",
    "20 mail2.example.com.",
  ]
}
```

### TXT Record

```hcl
# SPF Record
resource "bind9_record" "spf" {
  zone    = "example.com"
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 include:_spf.google.com ~all"]
}

# DKIM Record
resource "bind9_record" "dkim" {
  zone    = "example.com"
  name    = "google._domainkey"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4..."]
}

# DMARC Record
resource "bind9_record" "dmarc" {
  zone    = "example.com"
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"]
}

# Site Verification
resource "bind9_record" "google_verification" {
  zone    = "example.com"
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["google-site-verification=abc123xyz"]
}
```

### NS Record (Nameserver)

```hcl
resource "bind9_record" "ns" {
  zone    = "example.com"
  name    = "@"
  type    = "NS"
  ttl     = 86400
  records = [
    "ns1.example.com.",
    "ns2.example.com.",
  ]
}

# Delegate subdomain
resource "bind9_record" "subdomain_ns" {
  zone    = "example.com"
  name    = "subdomain"
  type    = "NS"
  ttl     = 86400
  records = [
    "ns1.subdomain.example.com.",
    "ns2.subdomain.example.com.",
  ]
}
```

### SRV Record (Service)

```hcl
# SIP Service
resource "bind9_record" "sip" {
  zone    = "example.com"
  name    = "_sip._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["10 60 5060 sip.example.com."]
}

# XMPP Service
resource "bind9_record" "xmpp_server" {
  zone    = "example.com"
  name    = "_xmpp-server._tcp"
  type    = "SRV"
  ttl     = 3600
  records = [
    "10 0 5269 xmpp1.example.com.",
    "20 0 5269 xmpp2.example.com.",
  ]
}

# LDAP Service
resource "bind9_record" "ldap" {
  zone    = "example.com"
  name    = "_ldap._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["0 100 389 ldap.example.com."]
}
```

### CAA Record (Certificate Authority)

```hcl
resource "bind9_record" "caa" {
  zone    = "example.com"
  name    = "@"
  type    = "CAA"
  ttl     = 3600
  records = [
    "0 issue \"letsencrypt.org\"",
    "0 issuewild \"letsencrypt.org\"",
    "0 iodef \"mailto:security@example.com\"",
  ]
}
```

### PTR Record (Reverse DNS)

```hcl
# Requires a reverse zone
resource "bind9_record" "ptr" {
  zone    = "1.168.192.in-addr.arpa"
  name    = "100"
  type    = "PTR"
  ttl     = 3600
  records = ["www.example.com."]
}
```

### NAPTR Record

```hcl
resource "bind9_record" "naptr" {
  zone    = "example.com"
  name    = "@"
  type    = "NAPTR"
  ttl     = 3600
  records = [
    "100 10 \"U\" \"E2U+sip\" \"!^.*$!sip:info@example.com!\" .",
  ]
}
```

### SSHFP Record (SSH Fingerprint)

```hcl
resource "bind9_record" "sshfp" {
  zone    = "example.com"
  name    = "server"
  type    = "SSHFP"
  ttl     = 3600
  records = [
    "1 1 abc123...",  # RSA SHA-1
    "1 2 def456...",  # RSA SHA-256
    "4 2 ghi789...",  # Ed25519 SHA-256
  ]
}
```

### TLSA Record (DANE)

```hcl
resource "bind9_record" "tlsa" {
  zone    = "example.com"
  name    = "_443._tcp.www"
  type    = "TLSA"
  ttl     = 3600
  records = ["3 1 1 abc123..."]  # DANE-EE, SPKI, SHA-256
}
```

## Argument Reference

### Required

- `zone` (String) - The zone name where the record will be created. Changing this forces a new resource.
- `name` (String) - The record name (e.g., `www`, `@` for zone apex, `_sip._tcp`). Changing this forces a new resource.
- `type` (String) - The record type. See [Supported Record Types](#supported-record-types). Changing this forces a new resource.
- `records` (List of String) - List of record values/data. Format depends on record type.

### Optional

- `ttl` (Number) - Time to live in seconds. Default: `3600`
- `class` (String) - Record class. Default: `IN`

### Read-Only

- `id` (String) - Record identifier in format `zone/name/type`.

## Supported Record Types

| Type | Description | Example Record Value |
|------|-------------|---------------------|
| `A` | IPv4 address | `192.168.1.100` |
| `AAAA` | IPv6 address | `2001:db8::1` |
| `CNAME` | Canonical name (alias) | `www.example.com.` |
| `MX` | Mail exchanger | `10 mail.example.com.` |
| `TXT` | Text record | `v=spf1 include:... ~all` |
| `NS` | Nameserver | `ns1.example.com.` |
| `PTR` | Pointer (reverse DNS) | `www.example.com.` |
| `SRV` | Service locator | `10 60 5060 sip.example.com.` |
| `CAA` | CA Authorization | `0 issue "letsencrypt.org"` |
| `NAPTR` | Naming Authority | `100 10 "U" "E2U+sip" "..." .` |
| `HTTPS` | HTTPS service binding | `1 . alpn="h2"` |
| `SVCB` | Service binding | `1 . alpn="h2"` |
| `TLSA` | TLS Authentication | `3 1 1 abc123...` |
| `SSHFP` | SSH fingerprint | `1 2 abc123...` |
| `DNSKEY` | DNSSEC public key | (auto-generated) |
| `DS` | Delegation signer | (auto-generated) |
| `LOC` | Geographic location | `37 46 30.000 N ...` |
| `HINFO` | Host information | `"x86_64" "Linux"` |
| `RP` | Responsible person | `admin.example.com. .` |
| `DNAME` | Delegation name | `other.example.com.` |
| `URI` | URI record | `10 1 "https://..."` |

## Record Value Formats

### MX Record
Format: `priority hostname`
```
10 mail.example.com.
```

### SRV Record
Format: `priority weight port target`
```
10 60 5060 sip.example.com.
```

### CAA Record
Format: `flags tag "value"`
```
0 issue "letsencrypt.org"
0 issuewild ";"
0 iodef "mailto:admin@example.com"
```

### TLSA Record
Format: `usage selector matching_type certificate_data`
```
3 1 1 abc123def456...
```

### SSHFP Record
Format: `algorithm fingerprint_type fingerprint`
```
1 2 abc123def456...
```

## Import

Records can be imported using the format `zone/name/type`:

```bash
terraform import bind9_record.www example.com/www/A
terraform import bind9_record.mx example.com/@/MX
terraform import bind9_record.sip example.com/_sip._tcp/SRV
```

## Notes

### Zone Apex (@)

Use `@` to refer to the zone apex (the domain itself):

```hcl
resource "bind9_record" "apex" {
  zone    = "example.com"
  name    = "@"
  type    = "A"
  records = ["192.168.1.1"]
}
```

### Trailing Dots

Hostnames in record data should typically end with a trailing dot to indicate they are fully qualified:

```hcl
# Correct - FQDN with trailing dot
records = ["mail.example.com."]

# Incorrect - will be interpreted as relative to zone
records = ["mail.example.com"]
```

### Multiple Records

You can create multiple records of the same type/name by including all values in the `records` list:

```hcl
resource "bind9_record" "multi_a" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  records = [
    "192.168.1.1",
    "192.168.1.2",
    "192.168.1.3",
  ]
}
```

### TTL Best Practices

| Use Case | Recommended TTL |
|----------|-----------------|
| High availability (load balancing) | 60-300 seconds |
| Standard web records | 300-3600 seconds |
| MX records | 3600-86400 seconds |
| NS records | 86400+ seconds |
| During migrations | 60-300 seconds |

