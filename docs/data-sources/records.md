---
page_title: "bind9_records Data Source - BIND9 Provider"
subcategory: "Record Management"
description: |-
  Retrieves all DNS records in a zone.
---

# bind9_records (Data Source)

Retrieves all DNS records from a zone with optional filtering by type or name.

## Example Usage

### List All Records in a Zone

```hcl
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

```hcl
# Get all A records
data "bind9_records" "a_records" {
  zone = "example.com"
  type = "A"
}

# Get all MX records
data "bind9_records" "mx_records" {
  zone = "example.com"
  type = "MX"
}

# Get all TXT records
data "bind9_records" "txt_records" {
  zone = "example.com"
  type = "TXT"
}

output "a_record_hosts" {
  value = [for r in data.bind9_records.a_records.records : r.name]
}
```

### Filter by Name

```hcl
# Get all records for a specific host
data "bind9_records" "www" {
  zone = "example.com"
  name = "www"
}

output "www_records" {
  value = data.bind9_records.www.records
}
```

### Combine Filters

```hcl
# Get A records for www only
data "bind9_records" "www_a" {
  zone = "example.com"
  name = "www"
  type = "A"
}
```

### Audit Zone Records

```hcl
data "bind9_records" "all" {
  zone = "example.com"
}

# Find all records with low TTL
output "low_ttl_records" {
  value = [for r in data.bind9_records.all.records : r if r.ttl < 300]
}

# Group records by type
output "records_by_type" {
  value = {
    for r in data.bind9_records.all.records :
    r.type => r.name...
  }
}
```

### Export Zone Data

```hcl
data "bind9_records" "all" {
  zone = "example.com"
}

output "zone_export" {
  value = [
    for r in data.bind9_records.all.records :
    "${r.name} ${r.ttl} IN ${r.type} ${r.rdata}"
  ]
}
```

## Argument Reference

### Required

- `zone` (String) - The zone name to query.

### Optional

- `type` (String) - Filter by record type (e.g., `A`, `AAAA`, `CNAME`, `MX`).
- `name` (String) - Filter by record name (e.g., `www`, `@`).

## Attribute Reference

The following attributes are exported:

- `id` (String) - Data source identifier.
- `zone` (String) - The zone name.
- `records` (List of Object) - List of DNS records. Each record has:
  - `name` (String) - Record name.
  - `type` (String) - Record type.
  - `ttl` (Number) - Record TTL.
  - `rdata` (String) - Record data/value.

## Use Cases

### Inventory Management

```hcl
data "bind9_records" "all_a" {
  zone = "example.com"
  type = "A"
}

# Create a map of hostnames to IPs
output "host_inventory" {
  value = {
    for r in data.bind9_records.all_a.records :
    r.name => r.rdata
  }
}
```

### Migration Planning

```hcl
data "bind9_records" "all" {
  zone = "old-domain.com"
}

# Use records to plan migration to new domain
resource "bind9_record" "migrated" {
  for_each = {
    for idx, r in data.bind9_records.all.records :
    "${r.name}-${r.type}-${idx}" => r
    if r.type != "SOA" && r.type != "NS"
  }
  
  zone    = "new-domain.com"
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = [each.value.rdata]
}
```

