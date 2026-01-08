---
page_title: "bind9_zones Data Source - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Retrieves a list of all DNS zones.
---

# bind9_zones (Data Source)

Retrieves a list of all DNS zones configured on the BIND9 server, with optional filtering.

## Example Usage

### List All Zones

```hcl
data "bind9_zones" "all" {}

output "zone_count" {
  value = length(data.bind9_zones.all.zones)
}

output "zone_names" {
  value = [for z in data.bind9_zones.all.zones : z.name]
}
```

### Filter by Zone Type

```hcl
# Get only master zones
data "bind9_zones" "masters" {
  type = "master"
}

# Get only slave zones
data "bind9_zones" "slaves" {
  type = "slave"
}

output "master_zones" {
  value = [for z in data.bind9_zones.masters.zones : z.name]
}
```

### Iterate Over Zones

```hcl
data "bind9_zones" "all" {}

# Create a CAA record for each zone
resource "bind9_record" "caa" {
  for_each = { for z in data.bind9_zones.all.zones : z.name => z if z.type == "master" }
  
  zone    = each.value.name
  name    = "@"
  type    = "CAA"
  ttl     = 3600
  records = ["0 issue \"letsencrypt.org\""]
}
```

## Argument Reference

### Optional

- `type` (String) - Filter zones by type. Valid values: `master`, `slave`, `forward`, `stub`.

## Attribute Reference

The following attributes are exported:

- `id` (String) - Data source identifier.
- `zones` (List of Object) - List of zones. Each zone has:
  - `id` (String) - Zone identifier.
  - `name` (String) - Zone name.
  - `type` (String) - Zone type.
  - `file` (String) - Zone file path.
  - `serial` (Number) - SOA serial number.
  - `loaded` (Boolean) - Whether zone is loaded.
  - `dnssec_enabled` (Boolean) - DNSSEC status.
  - `record_count` (Number) - Number of records.

