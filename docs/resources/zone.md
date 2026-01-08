---
page_title: "bind9_zone Resource - BIND9 Provider"
subcategory: "Zone Management"
description: |-
  Manages a DNS zone on BIND9 server.
---

# bind9_zone (Resource)

Manages a DNS zone on a BIND9 server. Supports master, slave, forward, and stub zone types with full control over SOA parameters, nameservers, glue records, and access control lists.

## Example Usage

### Basic Master Zone

```terraform
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
}
```

### Master Zone with Full Configuration

```terraform
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"

  # SOA Record Configuration
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 86400    # 1 day
  soa_retry   = 7200     # 2 hours
  soa_expire  = 3600000  # ~41 days
  soa_minimum = 3600     # 1 hour (negative cache TTL)

  # Default TTL for records
  default_ttl = 3600

  # Authoritative Nameservers
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]

  # Glue Records (required for in-zone nameservers)
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Access Control
  allow_transfer = ["10.0.0.0/8"]        # Zone transfer ACL
  allow_update   = ["key ddns-key"]       # Dynamic update ACL
  allow_query    = ["any"]                # Query ACL

  # Notifications
  notify = true

  # Cleanup behavior
  delete_file_on_destroy = false
}
```

### Slave Zone

```terraform
resource "bind9_zone" "slave" {
  name = "example.com"
  type = "slave"

  # Note: Master server configuration is done in BIND9 named.conf
}
```

### Forward Zone

```terraform
resource "bind9_zone" "forward" {
  name = "internal.example.com"
  type = "forward"

  # Forwarders are configured in BIND9 named.conf
}
```

### Reverse DNS Zone (IPv4)

```terraform
resource "bind9_zone" "reverse_ipv4" {
  name = "1.168.192.in-addr.arpa"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]

  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
}
```

### Reverse DNS Zone (IPv6)

```terraform
resource "bind9_zone" "reverse_ipv6" {
  name = "8.b.d.0.1.0.0.2.ip6.arpa"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
}
```

## Argument Reference

### Required

- `name` (String) The zone name (e.g., `example.com`, `1.168.192.in-addr.arpa`). **Changing this forces a new resource to be created.**
- `type` (String) The zone type. Valid values: `master`, `slave`, `forward`, `stub`. **Changing this forces a new resource to be created.**

### Optional

- `file` (String) Zone file path. If not specified, auto-generated based on zone name.
- `soa_mname` (String) Primary nameserver for SOA record. Default: `ns1`
- `soa_rname` (String) Responsible person email for SOA record (use `.` instead of `@`, e.g., `hostmaster.example.com`). Default: `hostmaster`
- `soa_refresh` (Number) SOA refresh interval in seconds. How often secondary servers should check for updates. Default: `86400` (1 day)
- `soa_retry` (Number) SOA retry interval in seconds. How long to wait before retrying a failed refresh. Default: `7200` (2 hours)
- `soa_expire` (Number) SOA expire time in seconds. When secondary servers should stop serving the zone if they can't reach the primary. Default: `3600000` (~41 days)
- `soa_minimum` (Number) SOA minimum/negative cache TTL in seconds. How long resolvers should cache NXDOMAIN responses. Default: `3600` (1 hour)
- `default_ttl` (Number) Default TTL for records in the zone. Default: `3600` (1 hour)
- `nameservers` (List of String) List of authoritative nameservers for the zone.
- `ns_addresses` (Map of String) Map of nameserver hostnames to IP addresses. **Required for in-zone nameservers** (glue records). Example: `{"ns1.example.com" = "10.0.1.10"}`
- `allow_transfer` (List of String) ACL for zone transfers (AXFR/IXFR). Examples: `["none"]`, `["10.0.0.0/8"]`, `["key transfer-key"]`
- `allow_update` (List of String) ACL for dynamic DNS updates. Examples: `["none"]`, `["key ddns-key"]`, `["10.0.1.0/24"]`
- `allow_query` (List of String) ACL for DNS queries. Examples: `["any"]`, `["10.0.0.0/8"]`, `["localhost"]`
- `notify` (Boolean) Send NOTIFY messages to slave servers when the zone changes. Default: `true`
- `delete_file_on_destroy` (Boolean) Delete the zone file when the zone resource is destroyed. Set to `false` in production to prevent accidental data loss. Default: `false`

### Read-Only

- `id` (String) The zone identifier (same as name).
- `serial` (Number) Current SOA serial number. Automatically incremented on zone changes.
- `loaded` (Boolean) Whether the zone is currently loaded in BIND9.
- `dnssec_enabled` (Boolean) Whether DNSSEC is enabled for this zone.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The zone name.
- `serial` - The current SOA serial number.
- `loaded` - Whether the zone is loaded in BIND9.
- `dnssec_enabled` - Whether the zone has DNSSEC enabled.

## Import

Zones can be imported using the zone name:

```bash
terraform import bind9_zone.example example.com
```

### Import Examples

```bash
# Import a forward zone
terraform import bind9_zone.myzone myzone.example.com

# Import a reverse DNS zone
terraform import bind9_zone.reverse 1.168.192.in-addr.arpa
```

## Notes

### Zone Types

| Type | Description | Use Case |
|------|-------------|----------|
| `master` | Primary authoritative zone | You manage the zone data directly |
| `slave` | Secondary zone (transfers from master) | Redundancy, load distribution |
| `forward` | Forward queries to other DNS servers | Internal delegation |
| `stub` | Like slave but only NS records | Delegation tracking |

### SOA Record Parameters

| Parameter | Purpose | Recommended Value |
|-----------|---------|-------------------|
| `soa_refresh` | How often secondaries check for updates | 3600-86400 (1h-1d) |
| `soa_retry` | Retry interval after failed refresh | 600-7200 (10m-2h) |
| `soa_expire` | When secondaries stop serving stale data | 604800-3600000 (1w-41d) |
| `soa_minimum` | Negative cache TTL (NXDOMAIN caching) | 60-3600 (1m-1h) |

### Access Control Lists (ACLs)

ACLs support various formats:

| Format | Example | Description |
|--------|---------|-------------|
| `any` | `["any"]` | Allow all |
| `none` | `["none"]` | Allow none |
| IP address | `["10.0.1.5"]` | Single host |
| CIDR | `["10.0.0.0/8"]` | Network range |
| TSIG key | `["key ddns-key"]` | Authenticated by key |
| ACL name | `["internal"]` | Named ACL from BIND9 config |

### Glue Records

When nameservers are within the zone they serve (e.g., `ns1.example.com` for zone `example.com`), you must provide their IP addresses via `ns_addresses`. This creates glue records that prevent circular dependencies.

```terraform
# ns1.example.com is IN the example.com zone, so we need glue records
ns_addresses = {
  "ns1.example.com" = "10.0.1.10"
  "ns2.example.com" = "10.0.1.11"
}
```

### Best Practices

1. **Always set meaningful SOA values** - `soa_mname` and `soa_rname` should be real hostnames
2. **Configure zone transfers securely** - Use TSIG keys or IP restrictions
3. **Use `notify = true`** for master zones to inform secondaries of changes
4. **Set `delete_file_on_destroy = false`** in production environments
5. **Use glue records** when nameservers are in-zone
