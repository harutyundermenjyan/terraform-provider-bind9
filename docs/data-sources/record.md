---
page_title: "bind9_record Data Source - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Retrieves DNS records by name and type.
---

# bind9_record (Data Source)

Retrieves DNS records from a zone by name and type.

## Example Usage

### Basic Usage

```hcl
data "bind9_record" "www" {
  zone = "example.com"
  name = "www"
  type = "A"
}

output "www_addresses" {
  value = data.bind9_record.www.records
}
```

### Get MX Records

```hcl
data "bind9_record" "mx" {
  zone = "example.com"
  name = "@"
  type = "MX"
}

output "mail_servers" {
  value = data.bind9_record.mx.records
}
```

### Get TXT Records (SPF)

```hcl
data "bind9_record" "spf" {
  zone = "example.com"
  name = "@"
  type = "TXT"
}

output "spf_record" {
  value = data.bind9_record.spf.records
}
```

### Use in Conditionals

```hcl
data "bind9_record" "existing" {
  zone = "example.com"
  name = "www"
  type = "A"
}

# Only create if TTL is high (indicating it might be safe to lower)
resource "bind9_record" "www_updated" {
  count = data.bind9_record.existing.ttl > 3600 ? 1 : 0
  
  zone    = "example.com"
  name    = "www"
  type    = "A"
  ttl     = 300
  records = data.bind9_record.existing.records
}
```

## Argument Reference

### Required

- `zone` (String) - The zone name.
- `name` (String) - The record name (e.g., `www`, `@`).
- `type` (String) - The record type (e.g., `A`, `AAAA`, `CNAME`, `MX`).

## Attribute Reference

The following attributes are exported:

- `id` (String) - Record identifier in format `zone/name/type`.
- `zone` (String) - The zone name.
- `name` (String) - The record name.
- `type` (String) - The record type.
- `ttl` (Number) - The record TTL.
- `records` (List of String) - The record values.

