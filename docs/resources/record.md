---
page_title: "bind9_record Resource - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Manages a DNS record on BIND9 server.
---

# bind9_record (Resource)

Manages DNS records on a BIND9 server. Supports all common DNS record types including A, AAAA, CNAME, MX, TXT, NS, PTR, SRV, CAA, and many more.

## Example Usage

### A Record (IPv4 Address)

```terraform
resource "bind9_record" "www" {
  zone    = "example.com"
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}
```

### A Record with Multiple IPs (Round-Robin)

```terraform
resource "bind9_record" "app" {
  zone    = "example.com"
  name    = "app"
  type    = "A"
  ttl     = 60
  records = [
    "10.0.1.101",
    "10.0.1.102",
    "10.0.1.103",
  ]
}
```

### AAAA Record (IPv6 Address)

```terraform
resource "bind9_record" "www_ipv6" {
  zone    = "example.com"
  name    = "www"
  type    = "AAAA"
  ttl     = 300
  records = ["2001:db8::1"]
}
```

### CNAME Record (Alias)

```terraform
resource "bind9_record" "blog" {
  zone    = "example.com"
  name    = "blog"
  type    = "CNAME"
  ttl     = 3600
  records = ["www.example.com."]  # Note the trailing dot for FQDN
}
```

### MX Record (Mail Exchange)

```terraform
resource "bind9_record" "mx" {
  zone    = "example.com"
  name    = "@"  # Zone apex
  type    = "MX"
  ttl     = 3600
  records = [
    "10 mail1.example.com.",   # Priority 10
    "20 mail2.example.com.",   # Priority 20 (backup)
  ]
}
```

### TXT Record (Text/SPF/DKIM)

```terraform
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
  records = ["v=DKIM1; k=rsa; p=MIGfMA0GCS..."]
}

# DMARC Record
resource "bind9_record" "dmarc" {
  zone    = "example.com"
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DMARC1; p=reject; rua=mailto:dmarc@example.com"]
}
```

### NS Record (Nameserver Delegation)

```terraform
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

### PTR Record (Reverse DNS)

```terraform
resource "bind9_record" "ptr_100" {
  zone    = "1.168.192.in-addr.arpa"
  name    = "100"  # For IP 192.168.1.100
  type    = "PTR"
  ttl     = 3600
  records = ["server.example.com."]
}

# IPv6 PTR Record
resource "bind9_record" "ptr_ipv6" {
  zone    = "8.b.d.0.1.0.0.2.ip6.arpa"
  name    = "1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0"
  type    = "PTR"
  ttl     = 3600
  records = ["server.example.com."]
}
```

### SRV Record (Service Location)

```terraform
# SIP over TCP
resource "bind9_record" "sip_tcp" {
  zone    = "example.com"
  name    = "_sip._tcp"
  type    = "SRV"
  ttl     = 3600
  records = [
    "10 60 5060 sip1.example.com.",  # priority weight port target
    "20 40 5060 sip2.example.com.",
  ]
}

# LDAP
resource "bind9_record" "ldap" {
  zone    = "example.com"
  name    = "_ldap._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["0 100 389 ldap.example.com."]
}

# Kerberos
resource "bind9_record" "kerberos" {
  zone    = "example.com"
  name    = "_kerberos._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["0 100 88 kdc.example.com."]
}
```

### CAA Record (Certificate Authority Authorization)

```terraform
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

### NAPTR Record (Name Authority Pointer)

```terraform
resource "bind9_record" "naptr" {
  zone    = "example.com"
  name    = "@"
  type    = "NAPTR"
  ttl     = 3600
  records = [
    "100 10 \"u\" \"E2U+sip\" \"!^.*$!sip:info@example.com!\" .",
  ]
}
```

### SSHFP Record (SSH Fingerprint)

```terraform
resource "bind9_record" "sshfp" {
  zone    = "example.com"
  name    = "server"
  type    = "SSHFP"
  ttl     = 3600
  records = [
    "1 1 abc123def456...",  # RSA SHA1
    "1 2 def456abc789...",  # RSA SHA256
    "4 2 789abc123def...",  # Ed25519 SHA256
  ]
}
```

### TLSA Record (DANE)

```terraform
resource "bind9_record" "tlsa" {
  zone    = "example.com"
  name    = "_443._tcp.www"
  type    = "TLSA"
  ttl     = 3600
  records = ["3 1 1 abc123def456..."]  # usage selector matching_type cert_data
}
```

### LOC Record (Geographic Location)

```terraform
resource "bind9_record" "loc" {
  zone    = "example.com"
  name    = "dc1"
  type    = "LOC"
  ttl     = 86400
  records = ["37 23 30.000 N 121 59 19.000 W 10.00m 100m 10m 10m"]
}
```

### DNAME Record (Delegation Name)

```terraform
resource "bind9_record" "dname" {
  zone    = "example.com"
  name    = "legacy"
  type    = "DNAME"
  ttl     = 3600
  records = ["newsite.example.com."]
}
```

### Wildcard Record

```terraform
resource "bind9_record" "wildcard" {
  zone    = "example.com"
  name    = "*"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}
```

### Zone Apex Record (@)

```terraform
resource "bind9_record" "apex" {
  zone    = "example.com"
  name    = "@"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}
```

## Argument Reference

### Required

- `zone` (String) The zone name where the record belongs. **Changing this forces a new resource to be created.**
- `name` (String) The record name (hostname). Use `@` for zone apex, `*` for wildcard. **Changing this forces a new resource to be created.**
- `type` (String) The record type. **Changing this forces a new resource to be created.** Supported types:
  - **Common:** `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `PTR`
  - **Service:** `SRV`, `NAPTR`, `URI`
  - **Security:** `CAA`, `TLSA`, `SSHFP`, `DNSKEY`, `DS`
  - **Modern:** `HTTPS`, `SVCB`
  - **Other:** `SOA`, `DNAME`, `LOC`, `HINFO`, `RP`
- `records` (List of String) The record data values. Format depends on record type (see examples above).

### Optional

- `ttl` (Number) Time to live in seconds. How long resolvers should cache this record. Default: `3600` (1 hour)
- `class` (String) Record class. Default: `IN`. Other values: `CH` (Chaosnet), `HS` (Hesiod).

### Convenience Attributes (Optional, Read-Only)

These attributes are automatically populated based on the record type and data:

- `address` (String) IP address for A/AAAA records.
- `target` (String) Target hostname for CNAME/NS/PTR/DNAME records.
- `priority` (Number) Priority value for MX/SRV records.
- `weight` (Number) Weight value for SRV records.
- `port` (Number) Port number for SRV records.
- `text` (String) Text content for TXT records.
- `flags` (Number) Flags value for CAA records.
- `tag` (String) Tag for CAA records (`issue`, `issuewild`, `iodef`).
- `value` (String) Value for CAA records.

### Read-Only

- `id` (String) The record identifier in format `zone/name/type`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The record identifier in format `zone/name/type`.
- Convenience attributes based on record type (see above).

## Import

Records can be imported using the format `zone/name/type`:

```bash
terraform import bind9_record.www example.com/www/A
```

### Import Examples

```bash
# Import an A record
terraform import bind9_record.www example.com/www/A

# Import a zone apex record
terraform import bind9_record.apex "example.com/@/A"

# Import an MX record
terraform import bind9_record.mx "example.com/@/MX"

# Import a wildcard record
terraform import bind9_record.wildcard "example.com/*/A"

# Import a PTR record
terraform import bind9_record.ptr "1.168.192.in-addr.arpa/100/PTR"
```

## Record Type Reference

### Record Format Guide

| Type | Format | Example |
|------|--------|---------|
| `A` | IPv4 address | `["10.0.1.100"]` |
| `AAAA` | IPv6 address | `["2001:db8::1"]` |
| `CNAME` | Target FQDN (with trailing dot) | `["www.example.com."]` |
| `MX` | `priority target` | `["10 mail.example.com."]` |
| `TXT` | Text string | `["v=spf1 -all"]` |
| `NS` | Nameserver FQDN | `["ns1.example.com."]` |
| `PTR` | Target FQDN | `["server.example.com."]` |
| `SRV` | `priority weight port target` | `["10 60 5060 sip.example.com."]` |
| `CAA` | `flags tag value` | `["0 issue \"letsencrypt.org\""]` |
| `SSHFP` | `algorithm fptype fingerprint` | `["1 1 abc123..."]` |
| `TLSA` | `usage selector matching_type data` | `["3 1 1 abc123..."]` |

### TTL Best Practices

| Record Type | Recommended TTL | Reason |
|-------------|-----------------|--------|
| A/AAAA | 300-3600 | Balance between caching and flexibility |
| CNAME | 3600+ | Aliases rarely change |
| MX | 3600-86400 | Mail routing should be stable |
| NS | 86400+ | Delegation changes are rare |
| TXT (SPF/DKIM) | 3600+ | Security records should be cached |
| PTR | 3600-86400 | Reverse DNS is typically stable |
| SRV | 300-3600 | Service discovery needs accuracy |
| SOA minimum | 60-3600 | Affects negative caching |

### Important Notes

1. **Trailing dots for FQDNs** - Always use trailing dots for fully qualified domain names in CNAME, MX, NS, PTR, SRV targets (e.g., `mail.example.com.`). Without the trailing dot, BIND9 appends the zone name.

2. **Zone apex restrictions** - You cannot create a CNAME at the zone apex (`@`). Use A/AAAA records instead, or consider ALIAS/ANAME if supported.

3. **Record ordering** - When creating both CNAME and other records for the same name, be aware that CNAME records cannot coexist with other record types for the same name.

4. **Multiple records** - The `records` list can contain multiple values for round-robin (A/AAAA) or failover (MX with priorities).

5. **Escaping in TXT records** - Long TXT records or records with special characters may need escaping. The provider handles most cases automatically.
