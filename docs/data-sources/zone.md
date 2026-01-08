---
page_title: "bind9_zone Data Source - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Retrieves information about a DNS zone.
---

# bind9_zone (Data Source)

Retrieves information about a specific DNS zone on the BIND9 server. Use this data source to query zone properties like serial number, loaded status, and DNSSEC status.

## Example Usage

### Basic Usage

```terraform
data "bind9_zone" "example" {
  name = "example.com"
}

output "zone_serial" {
  value = data.bind9_zone.example.serial
}
```

### Check Zone Status

```terraform
data "bind9_zone" "production" {
  name = "production.example.com"
}

output "zone_info" {
  value = {
    name           = data.bind9_zone.production.name
    type           = data.bind9_zone.production.type
    serial         = data.bind9_zone.production.serial
    loaded         = data.bind9_zone.production.loaded
    dnssec_enabled = data.bind9_zone.production.dnssec_enabled
    record_count   = data.bind9_zone.production.record_count
  }
}
```

### Use Zone Data in Records

```terraform
data "bind9_zone" "main" {
  name = "example.com"
}

resource "bind9_record" "www" {
  zone    = data.bind9_zone.main.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}
```

### Conditional Logic Based on Zone State

```terraform
data "bind9_zone" "check" {
  name = "example.com"
}

# Only create records if zone is loaded
resource "bind9_record" "app" {
  count = data.bind9_zone.check.loaded ? 1 : 0

  zone    = data.bind9_zone.check.name
  name    = "app"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.101"]
}
```

### Check DNSSEC Status

```terraform
data "bind9_zone" "secure" {
  name = "secure.example.com"
}

output "dnssec_status" {
  value = data.bind9_zone.secure.dnssec_enabled ? "DNSSEC is enabled" : "DNSSEC is NOT enabled"
}
```

## Argument Reference

### Required

- `name` (String) The zone name to query (e.g., `example.com`, `1.168.192.in-addr.arpa`).

## Attribute Reference

The following attributes are exported:

- `id` (String) The zone identifier (same as name).
- `name` (String) The zone name.
- `type` (String) The zone type (`master`, `slave`, `forward`, `stub`).
- `file` (String) The zone file path on the BIND9 server.
- `serial` (Number) The current SOA serial number.
- `loaded` (Boolean) Whether the zone is currently loaded in BIND9.
- `dnssec_enabled` (Boolean) Whether DNSSEC is enabled for this zone.
- `record_count` (Number) The total number of records in the zone.

## Use Cases

### Pre-flight Checks

Use the zone data source to verify a zone exists and is loaded before creating records:

```terraform
data "bind9_zone" "target" {
  name = var.target_zone
}

# This will fail during plan if zone doesn't exist
output "zone_verified" {
  value = "Zone ${data.bind9_zone.target.name} is ready (serial: ${data.bind9_zone.target.serial})"
}
```

### Cross-Zone References

Reference zones dynamically without hardcoding:

```terraform
variable "zones" {
  default = ["example.com", "example.org"]
}

data "bind9_zone" "all" {
  for_each = toset(var.zones)
  name     = each.value
}

output "all_zone_serials" {
  value = {
    for name, zone in data.bind9_zone.all : name => zone.serial
  }
}
```

### DNSSEC Validation

Ensure DNSSEC is properly configured before adding sensitive records:

```terraform
data "bind9_zone" "dnssec_check" {
  name = "secure.example.com"
}

resource "bind9_record" "sensitive" {
  count = data.bind9_zone.dnssec_check.dnssec_enabled ? 1 : 0

  zone    = data.bind9_zone.dnssec_check.name
  name    = "api"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.200"]

  lifecycle {
    precondition {
      condition     = data.bind9_zone.dnssec_check.dnssec_enabled
      error_message = "DNSSEC must be enabled for sensitive records."
    }
  }
}
```
