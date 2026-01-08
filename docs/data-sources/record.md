---
page_title: "bind9_record Data Source - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Retrieves DNS record(s) by name and type.
---

# bind9_record (Data Source)

Retrieves specific DNS records from a zone by name and type. Use this data source to query existing record values, verify configurations, or use record data in other resources.

## Example Usage

### Basic Usage

```terraform
data "bind9_record" "www" {
  zone = "example.com"
  name = "www"
  type = "A"
}

output "www_ips" {
  value = data.bind9_record.www.records
}
```

### Query MX Records

```terraform
data "bind9_record" "mx" {
  zone = "example.com"
  name = "@"
  type = "MX"
}

output "mail_servers" {
  value = data.bind9_record.mx.records
}
```

### Query TXT Records (SPF)

```terraform
data "bind9_record" "spf" {
  zone = "example.com"
  name = "@"
  type = "TXT"
}

output "spf_record" {
  value = data.bind9_record.spf.records
}
```

### Query Nameservers

```terraform
data "bind9_record" "ns" {
  zone = "example.com"
  name = "@"
  type = "NS"
}

output "nameservers" {
  value = data.bind9_record.ns.records
}
```

### Query PTR Record

```terraform
data "bind9_record" "ptr" {
  zone = "1.168.192.in-addr.arpa"
  name = "100"
  type = "PTR"
}

output "reverse_dns" {
  value = data.bind9_record.ptr.records[0]
}
```

### Use Record Data in Another Resource

```terraform
# Get existing load balancer IPs
data "bind9_record" "lb" {
  zone = "example.com"
  name = "lb"
  type = "A"
}

# Create alias pointing to the load balancer
resource "bind9_record" "app" {
  zone    = "example.com"
  name    = "app"
  type    = "CNAME"
  ttl     = 300
  records = ["lb.example.com."]
}
```

### Validate Record Existence

```terraform
data "bind9_record" "required" {
  zone = "example.com"
  name = "www"
  type = "A"
}

resource "null_resource" "validate" {
  # This will fail if www.example.com A record doesn't exist
  triggers = {
    record_exists = length(data.bind9_record.required.records) > 0 ? "true" : ""
  }
}
```

### Check Record TTL

```terraform
data "bind9_record" "check_ttl" {
  zone = "example.com"
  name = "www"
  type = "A"
}

output "ttl_info" {
  value = {
    name    = data.bind9_record.check_ttl.name
    type    = data.bind9_record.check_ttl.type
    ttl     = data.bind9_record.check_ttl.ttl
    records = data.bind9_record.check_ttl.records
  }
}
```

## Argument Reference

### Required

- `zone` (String) The zone name to query.
- `name` (String) The record name to query. Use `@` for zone apex, `*` for wildcard.
- `type` (String) The record type to query (e.g., `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `PTR`, `SRV`, `CAA`).

## Attribute Reference

The following attributes are exported:

- `id` (String) The record identifier in format `zone/name/type`.
- `zone` (String) The zone name.
- `name` (String) The record name.
- `type` (String) The record type.
- `ttl` (Number) The record TTL in seconds.
- `records` (List of String) The record values.

## Error Handling

The data source will return an error if:

1. The zone doesn't exist
2. No records match the specified name and type
3. The BIND9 API is unreachable

```terraform
# This will fail during plan if record doesn't exist
data "bind9_record" "must_exist" {
  zone = "example.com"
  name = "www"
  type = "A"
}
```

## Use Cases

### DNS Validation

```terraform
# Verify that required DNS records exist
data "bind9_record" "validate_a" {
  zone = "example.com"
  name = "www"
  type = "A"
}

data "bind9_record" "validate_mx" {
  zone = "example.com"
  name = "@"
  type = "MX"
}

output "dns_validation" {
  value = {
    www_exists    = length(data.bind9_record.validate_a.records) > 0
    www_ips       = data.bind9_record.validate_a.records
    mx_configured = length(data.bind9_record.validate_mx.records) > 0
    mail_servers  = data.bind9_record.validate_mx.records
  }
}
```

### Cross-Reference Records

```terraform
# Check if the target of a CNAME exists
data "bind9_record" "cname_target" {
  zone = "example.com"
  name = "www"
  type = "A"
}

resource "bind9_record" "alias" {
  zone    = "example.com"
  name    = "site"
  type    = "CNAME"
  ttl     = 300
  records = ["www.example.com."]

  lifecycle {
    precondition {
      condition     = length(data.bind9_record.cname_target.records) > 0
      error_message = "Cannot create CNAME - target www.example.com has no A records."
    }
  }
}
```

### Compare with Expected Values

```terraform
variable "expected_ips" {
  default = ["10.0.1.100", "10.0.1.101"]
}

data "bind9_record" "current" {
  zone = "example.com"
  name = "www"
  type = "A"
}

locals {
  ips_match = toset(data.bind9_record.current.records) == toset(var.expected_ips)
}

output "ip_verification" {
  value = {
    expected = var.expected_ips
    actual   = data.bind9_record.current.records
    match    = local.ips_match
  }
}
```
