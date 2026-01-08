---
page_title: "bind9_zones Data Source - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Retrieves list of all DNS zones.
---

# bind9_zones (Data Source)

Retrieves a list of all DNS zones on the BIND9 server with optional filtering by zone type. Use this data source to discover zones, audit configuration, or iterate over multiple zones.

## Example Usage

### List All Zones

```terraform
data "bind9_zones" "all" {}

output "zone_count" {
  value = length(data.bind9_zones.all.zones)
}

output "all_zones" {
  value = [for z in data.bind9_zones.all.zones : z.name]
}
```

### Filter by Zone Type

```terraform
# Get only master zones
data "bind9_zones" "masters" {
  type = "master"
}

output "master_zones" {
  value = [for z in data.bind9_zones.masters.zones : z.name]
}

# Get slave zones
data "bind9_zones" "slaves" {
  type = "slave"
}
```

### Zone Inventory Report

```terraform
data "bind9_zones" "all" {}

output "zone_inventory" {
  value = {
    for z in data.bind9_zones.all.zones : z.name => {
      type           = z.type
      serial         = z.serial
      loaded         = z.loaded
      dnssec_enabled = z.dnssec_enabled
      record_count   = z.record_count
    }
  }
}
```

### Find Zones Not Yet Signed

```terraform
data "bind9_zones" "masters" {
  type = "master"
}

output "unsigned_zones" {
  value = [
    for z in data.bind9_zones.masters.zones : z.name
    if !z.dnssec_enabled
  ]
}
```

### Iterate Over Zones to Create Records

```terraform
data "bind9_zones" "masters" {
  type = "master"
}

# Create a monitoring record in each master zone
resource "bind9_record" "monitoring" {
  for_each = {
    for z in data.bind9_zones.masters.zones : z.name => z
    if z.loaded
  }

  zone    = each.value.name
  name    = "_health"
  type    = "TXT"
  ttl     = 60
  records = ["monitoring-enabled"]
}
```

### Zone Statistics

```terraform
data "bind9_zones" "all" {}

locals {
  zones_by_type = {
    for z in data.bind9_zones.all.zones : z.type => z.name...
  }
  
  total_records = sum([
    for z in data.bind9_zones.all.zones : z.record_count
  ])
  
  dnssec_zones = length([
    for z in data.bind9_zones.all.zones : z.name
    if z.dnssec_enabled
  ])
}

output "dns_stats" {
  value = {
    total_zones     = length(data.bind9_zones.all.zones)
    master_zones    = try(length(local.zones_by_type["master"]), 0)
    slave_zones     = try(length(local.zones_by_type["slave"]), 0)
    total_records   = local.total_records
    dnssec_enabled  = local.dnssec_zones
    dnssec_coverage = "${local.dnssec_zones}/${length(data.bind9_zones.all.zones)}"
  }
}
```

### Check for Unloaded Zones

```terraform
data "bind9_zones" "all" {}

output "problem_zones" {
  value = [
    for z in data.bind9_zones.all.zones : {
      name   = z.name
      type   = z.type
      file   = z.file
      loaded = z.loaded
    }
    if !z.loaded
  ]
  description = "Zones that are not currently loaded (may indicate configuration errors)"
}
```

## Argument Reference

### Optional

- `type` (String) Filter zones by type. Valid values: `master`, `slave`, `forward`, `stub`. If not specified, returns all zones.

## Attribute Reference

The following attributes are exported:

- `id` (String) The data source identifier (always "zones").
- `zones` (List of Object) List of zone objects. Each zone has:
  - `id` (String) Zone identifier (same as name).
  - `name` (String) Zone name.
  - `type` (String) Zone type (`master`, `slave`, `forward`, `stub`).
  - `file` (String) Zone file path on the server.
  - `serial` (Number) Current SOA serial number.
  - `loaded` (Boolean) Whether zone is loaded.
  - `dnssec_enabled` (Boolean) Whether DNSSEC is enabled.
  - `record_count` (Number) Number of records in zone.

## Use Cases

### Configuration Drift Detection

```terraform
variable "expected_zones" {
  type = set(string)
  default = [
    "example.com",
    "example.org",
    "internal.example.com",
  ]
}

data "bind9_zones" "masters" {
  type = "master"
}

locals {
  actual_zones = toset([for z in data.bind9_zones.masters.zones : z.name])
  missing      = setsubtract(var.expected_zones, local.actual_zones)
  unexpected   = setsubtract(local.actual_zones, var.expected_zones)
}

output "configuration_drift" {
  value = {
    missing_zones    = local.missing
    unexpected_zones = local.unexpected
    all_expected     = length(local.missing) == 0 && length(local.unexpected) == 0
  }
}
```

### Bulk Operations

```terraform
data "bind9_zones" "to_update" {
  type = "master"
}

# Add SPF record to all zones
resource "bind9_record" "spf" {
  for_each = {
    for z in data.bind9_zones.to_update.zones : z.name => z
  }

  zone    = each.value.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 include:_spf.example.com ~all"]
}
```

### Zone Backup Verification

```terraform
data "bind9_zones" "all" {}

output "zone_file_locations" {
  value = {
    for z in data.bind9_zones.all.zones : z.name => z.file
    if z.file != null && z.file != ""
  }
  description = "Zone file paths for backup purposes"
}
```
