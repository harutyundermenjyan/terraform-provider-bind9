---
page_title: "bind9_records Data Source - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Retrieves all DNS records in a zone.
---

# bind9_records (Data Source)

Retrieves all DNS records from a zone with optional filtering by record type or name. Use this data source to audit zone contents, export configurations, or perform bulk operations.

## Example Usage

### Get All Records in a Zone

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

output "record_count" {
  value = length(data.bind9_records.all.records)
}

output "all_records" {
  value = data.bind9_records.all.records
}
```

### Filter by Record Type

```terraform
# Get all A records
data "bind9_records" "a_records" {
  zone = "example.com"
  type = "A"
}

output "a_records" {
  value = [
    for r in data.bind9_records.a_records.records : {
      name  = r.name
      value = r.rdata
      ttl   = r.ttl
    }
  ]
}

# Get all CNAME records
data "bind9_records" "cnames" {
  zone = "example.com"
  type = "CNAME"
}
```

### Filter by Name

```terraform
# Get all record types for 'www'
data "bind9_records" "www_all" {
  zone = "example.com"
  name = "www"
}

output "www_records" {
  value = {
    for r in data.bind9_records.www_all.records : "${r.type}" => r.rdata
  }
}
```

### Combine Filters

```terraform
# Get A records for 'www' only
data "bind9_records" "www_a" {
  zone = "example.com"
  name = "www"
  type = "A"
}
```

### Zone Audit Report

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

locals {
  records_by_type = {
    for r in data.bind9_records.all.records : r.type => r...
  }
}

output "zone_audit" {
  value = {
    zone          = data.bind9_records.all.zone
    total_records = length(data.bind9_records.all.records)
    by_type       = { for t, recs in local.records_by_type : t => length(recs) }
  }
}
```

### Find Records with Low TTL

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

output "low_ttl_records" {
  value = [
    for r in data.bind9_records.all.records : {
      name = r.name
      type = r.type
      ttl  = r.ttl
    }
    if r.ttl < 300
  ]
  description = "Records with TTL less than 5 minutes"
}
```

### Export Zone as Table

```terraform
data "bind9_records" "export" {
  zone = "example.com"
}

output "zone_export" {
  value = [
    for r in data.bind9_records.export.records : 
    "${r.name}\t${r.ttl}\tIN\t${r.type}\t${r.rdata}"
  ]
}
```

### Compare Two Zones

```terraform
data "bind9_records" "zone1" {
  zone = "example.com"
}

data "bind9_records" "zone2" {
  zone = "example.org"
}

locals {
  zone1_names = toset([for r in data.bind9_records.zone1.records : r.name])
  zone2_names = toset([for r in data.bind9_records.zone2.records : r.name])
}

output "zone_comparison" {
  value = {
    only_in_zone1    = setsubtract(local.zone1_names, local.zone2_names)
    only_in_zone2    = setsubtract(local.zone2_names, local.zone1_names)
    in_both          = setintersection(local.zone1_names, local.zone2_names)
    zone1_count      = length(data.bind9_records.zone1.records)
    zone2_count      = length(data.bind9_records.zone2.records)
  }
}
```

### Find Duplicate Records

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

locals {
  record_keys = [
    for r in data.bind9_records.all.records : "${r.name}/${r.type}"
  ]
  
  record_counts = {
    for key in distinct(local.record_keys) : key => length([
      for k in local.record_keys : k if k == key
    ])
  }
  
  duplicates = {
    for key, count in local.record_counts : key => count
    if count > 1
  }
}

output "records_with_multiple_values" {
  value = local.duplicates
}
```

## Argument Reference

### Required

- `zone` (String) The zone name to query.

### Optional

- `type` (String) Filter by record type (e.g., `A`, `AAAA`, `CNAME`, `MX`, `TXT`).
- `name` (String) Filter by record name.

## Attribute Reference

The following attributes are exported:

- `id` (String) The data source identifier (same as zone name).
- `zone` (String) The zone name.
- `type` (String) The filter type (if specified).
- `name` (String) The filter name (if specified).
- `records` (List of Object) List of record objects. Each record has:
  - `name` (String) Record name.
  - `type` (String) Record type.
  - `ttl` (Number) Record TTL in seconds.
  - `rdata` (String) Record data value.

## Use Cases

### Security Audit

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

output "security_audit" {
  value = {
    spf_configured = length([
      for r in data.bind9_records.all.records : r
      if r.type == "TXT" && startswith(r.rdata, "v=spf1")
    ]) > 0
    
    dmarc_configured = length([
      for r in data.bind9_records.all.records : r
      if r.type == "TXT" && r.name == "_dmarc" && startswith(r.rdata, "v=DMARC1")
    ]) > 0
    
    caa_configured = length([
      for r in data.bind9_records.all.records : r
      if r.type == "CAA"
    ]) > 0
    
    wildcard_records = [
      for r in data.bind9_records.all.records : "${r.type}: ${r.rdata}"
      if r.name == "*"
    ]
  }
}
```

### Generate Terraform Import Commands

```terraform
data "bind9_records" "all" {
  zone = "example.com"
}

output "import_commands" {
  value = [
    for r in data.bind9_records.all.records :
    "terraform import 'bind9_record.${replace(r.name, ".", "_")}_${lower(r.type)}' '${data.bind9_records.all.zone}/${r.name}/${r.type}'"
    if r.type != "SOA" && r.type != "NS" || r.name != "@"
  ]
}
```

### Backup Zone Configuration

```terraform
data "bind9_records" "backup" {
  zone = "example.com"
}

# Output as JSON for backup
output "zone_backup_json" {
  value = jsonencode({
    zone      = data.bind9_records.backup.zone
    exported  = timestamp()
    records   = data.bind9_records.backup.records
  })
}
```

### Find Records Pointing to Specific IP

```terraform
variable "target_ip" {
  default = "10.0.1.100"
}

data "bind9_records" "a_records" {
  zone = "example.com"
  type = "A"
}

output "records_using_ip" {
  value = [
    for r in data.bind9_records.a_records.records : r.name
    if r.rdata == var.target_ip
  ]
}
```
