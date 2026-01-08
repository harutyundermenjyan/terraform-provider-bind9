---
page_title: "bind9_zone Data Source - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Retrieves information about a DNS zone.
---

# bind9_zone (Data Source)

Retrieves information about an existing DNS zone on the BIND9 server.

## Example Usage

### Basic Usage

```hcl
data "bind9_zone" "example" {
  name = "example.com"
}

output "zone_serial" {
  value = data.bind9_zone.example.serial
}

output "zone_type" {
  value = data.bind9_zone.example.type
}
```

### Check DNSSEC Status

```hcl
data "bind9_zone" "example" {
  name = "example.com"
}

output "dnssec_enabled" {
  value = data.bind9_zone.example.dnssec_enabled
}
```

### Use in Other Resources

```hcl
data "bind9_zone" "existing" {
  name = "example.com"
}

# Add a record to an existing zone
resource "bind9_record" "new_host" {
  zone    = data.bind9_zone.existing.name
  name    = "newhost"
  type    = "A"
  ttl     = 300
  records = ["192.168.1.200"]
}
```

## Argument Reference

### Required

- `name` (String) - The zone name to look up.

## Attribute Reference

The following attributes are exported:

- `id` (String) - Zone identifier (same as name).
- `name` (String) - The zone name.
- `type` (String) - Zone type (`master`, `slave`, `forward`, `stub`).
- `file` (String) - Path to the zone file.
- `serial` (Number) - Current SOA serial number.
- `loaded` (Boolean) - Whether the zone is currently loaded.
- `dnssec_enabled` (Boolean) - Whether DNSSEC is enabled for this zone.
- `record_count` (Number) - Number of records in the zone.

