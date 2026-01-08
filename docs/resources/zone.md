---
page_title: "bind9_zone Resource - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Manages a DNS zone on BIND9 server.
---

# bind9_zone (Resource)

Manages a DNS zone on a BIND9 server. Supports master, slave, forward, and stub zone types.

## Example Usage

### Basic Master Zone

```hcl
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
}
```

### Master Zone with Full Configuration

```hcl
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
  
  # SOA record configuration
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 86400   # 1 day
  soa_retry   = 7200    # 2 hours
  soa_expire  = 3600000 # ~41 days
  soa_minimum = 3600    # 1 hour
  
  # Default TTL for records
  default_ttl = 3600
  
  # Authoritative nameservers
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
  
  # Zone transfer settings
  allow_transfer = ["192.168.1.0/24", "10.0.0.0/8"]
  allow_update   = ["key rndc-key"]
  allow_query    = ["any"]
  
  # Send NOTIFY to slaves
  notify = true
  
  # Delete zone file when resource is destroyed
  delete_file_on_destroy = false
}
```

### Slave Zone

```hcl
resource "bind9_zone" "slave" {
  name = "example.com"
  type = "slave"
  
  # Note: Masters are configured in BIND9 directly
  # or via the options block
}
```

### Forward Zone

```hcl
resource "bind9_zone" "forward" {
  name = "internal.example.com"
  type = "forward"
  
  # Forwarders would be configured in BIND9
}
```

### Reverse DNS Zone

```hcl
resource "bind9_zone" "reverse" {
  name = "1.168.192.in-addr.arpa"
  type = "master"
  
  soa_mname = "ns1.example.com"
  soa_rname = "hostmaster.example.com"
  
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
}
```

## Argument Reference

### Required

- `name` (String) - The zone name (e.g., `example.com`, `1.168.192.in-addr.arpa`). Changing this forces a new resource to be created.
- `type` (String) - The zone type. Valid values: `master`, `slave`, `forward`, `stub`. Changing this forces a new resource to be created.

### Optional

- `file` (String) - Zone file path. If not specified, auto-generated based on zone name.
- `soa_mname` (String) - Primary nameserver for SOA record. Default: `ns1`
- `soa_rname` (String) - Responsible person email for SOA (use `.` instead of `@`). Default: `hostmaster`
- `soa_refresh` (Number) - SOA refresh interval in seconds. Default: `86400`
- `soa_retry` (Number) - SOA retry interval in seconds. Default: `7200`
- `soa_expire` (Number) - SOA expire time in seconds. Default: `3600000`
- `soa_minimum` (Number) - SOA minimum/negative TTL in seconds. Default: `3600`
- `default_ttl` (Number) - Default TTL for records in seconds. Default: `3600`
- `nameservers` (List of String) - List of authoritative nameservers for the zone.
- `allow_transfer` (List of String) - ACL for zone transfers. Default: `["none"]`
- `allow_update` (List of String) - ACL for dynamic updates. Default: `["none"]`
- `allow_query` (List of String) - ACL for queries. Default: `["any"]`
- `notify` (Boolean) - Send NOTIFY to slaves on zone changes. Default: `true`
- `delete_file_on_destroy` (Boolean) - Delete zone file when the zone is destroyed. Default: `false`

### Read-Only

- `id` (String) - The zone identifier (same as name).
- `serial` (Number) - Current zone serial number.
- `loaded` (Boolean) - Whether the zone is currently loaded.
- `dnssec_enabled` (Boolean) - Whether DNSSEC is enabled for this zone.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `serial` - The current SOA serial number.
- `loaded` - Whether the zone is loaded in BIND9.
- `dnssec_enabled` - Whether the zone has DNSSEC enabled.

## Import

Zones can be imported using the zone name:

```bash
terraform import bind9_zone.example example.com
```

### Import Example

```bash
# Import an existing zone
terraform import bind9_zone.myzone myzone.example.com

# Import a reverse DNS zone
terraform import bind9_zone.reverse 1.168.192.in-addr.arpa
```

## Notes

### SOA Serial Number

The SOA serial number is automatically managed by BIND9 when using dynamic updates. If you're managing zones through Terraform, the serial will be updated automatically on changes.

### Zone Types

| Type | Description |
|------|-------------|
| `master` | Primary authoritative zone (you manage the zone file) |
| `slave` | Secondary zone (transfers from master) |
| `forward` | Forward queries to other DNS servers |
| `stub` | Like slave but only NS records |

### Best Practices

1. Always set meaningful `soa_mname` and `soa_rname` values
2. Configure `allow_transfer` to restrict zone transfers
3. Use `notify = true` to inform slave servers of changes
4. Set `delete_file_on_destroy = false` in production to prevent accidental data loss

