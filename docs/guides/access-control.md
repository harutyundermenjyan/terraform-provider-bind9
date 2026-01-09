---
page_title: "Access Control Lists (ACLs) Guide"
subcategory: "Guides"
description: |-
  Learn how to configure access control for DNS zones including allow_query, allow_transfer, and allow_update.
---

# Access Control Lists (ACLs) Guide

This guide explains how to configure access control for DNS zones using `allow_query`, `allow_transfer`, and `allow_update` options.

## Overview

BIND9 provides three main access control options for zones:

| ACL Option | Purpose | Default |
|------------|---------|---------|
| `allow_query` | Who can query records in this zone | `any` |
| `allow_transfer` | Who can perform zone transfers (AXFR/IXFR) | `none` |
| `allow_update` | Who can dynamically update records | `none` |

## ACL Formats

The BIND9 provider supports all standard BIND9 ACL formats:

| Format | Example | Description |
|--------|---------|-------------|
| Keywords | `"any"`, `"none"`, `"localhost"`, `"localnets"` | Built-in ACL keywords |
| IP Address | `"10.0.1.5"` | Single host |
| CIDR Network | `"10.0.0.0/8"`, `"192.168.1.0/24"` | Network range |
| TSIG Key | `"key ddns-key"` | Authenticated by TSIG key |
| Named ACL | `"internal"`, `"trusted-secondaries"` | Custom ACL defined in named.conf |
| Negation | `"!10.0.1.99"` | Exclude specific address |

## Basic Examples

### Public Zone (Allow All Queries)

```terraform
resource "bind9_zone" "public" {
  name = "example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Public zone - anyone can query
  allow_query = ["any"]
  
  # Only allow zone transfers to secondary nameserver
  allow_transfer = ["10.0.1.11"]
  
  # No dynamic updates
  allow_update = ["none"]
  
  notify = true
}
```

### Internal Zone (Restricted Queries)

```terraform
resource "bind9_zone" "internal" {
  name = "internal.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Only internal networks can query
  allow_query = [
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
    "localhost",
  ]
  
  # Only specific secondary can transfer
  allow_transfer = ["10.0.1.11"]
  
  # Allow updates from DHCP server
  allow_update = ["10.0.1.5"]
  
  notify = true
}
```

### Zone with TSIG Key Authentication

```terraform
resource "bind9_zone" "secure" {
  name = "secure.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "security.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Internal queries only
  allow_query = ["10.0.0.0/8"]
  
  # Zone transfers require TSIG authentication
  allow_transfer = ["key transfer-key"]
  
  # Dynamic updates require TSIG authentication
  allow_update = ["key ddns-key"]
  
  notify = true
}
```

## Using Named ACLs

You can reference named ACLs defined in your BIND9 `named.conf`:

### Step 1: Define ACLs in named.conf

```bind
// /etc/bind/named.conf or /etc/bind/named.conf.options

acl "internal" {
    localhost;
    localnets;
    10.0.0.0/8;
    172.16.0.0/12;
    192.168.0.0/16;
};

acl "trusted-secondaries" {
    10.0.1.11;        // ns2
    10.0.1.12;        // ns3
    key "transfer-key";
};

acl "ddns-clients" {
    10.0.2.0/24;      // DHCP server network
    key "ddns-key";
};

acl "dmz" {
    10.0.100.0/24;
};
```

### Step 2: Reference ACLs in Terraform

```terraform
resource "bind9_zone" "using_named_acls" {
  name = "example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Use named ACL from named.conf
  allow_query = ["internal"]
  
  # Use named ACL for zone transfers
  allow_transfer = ["trusted-secondaries"]
  
  # Use named ACL for dynamic updates
  allow_update = ["ddns-clients"]
  
  notify = true
}
```

## Advanced ACL Patterns

### Multiple ACL Entries

```terraform
resource "bind9_zone" "multi_acl" {
  name = "example.com"
  type = "master"

  # ... SOA and nameserver config ...

  # Combine multiple formats
  allow_query = [
    "any",
  ]
  
  allow_transfer = [
    "10.0.1.11",              # Secondary NS by IP
    "10.0.1.12",              # Another secondary
    "key transfer-key",        # TSIG-authenticated transfers
  ]
  
  allow_update = [
    "10.0.2.5",               # DHCP server
    "key ddns-key",            # Authenticated updates
    "key dhcp-key",            # Another key
  ]
}
```

### DMZ Zone Configuration

```terraform
resource "bind9_zone" "dmz" {
  name = "dmz.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 600  # Shorter TTL for DMZ

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # DMZ accessible from internal and DMZ networks
  allow_query = [
    "internal",
    "dmz",
  ]
  
  # Local network transfers only
  allow_transfer = ["localnets"]
  
  # Updates from automation server
  allow_update = ["key ddns-key"]
  
  notify = true
}
```

### Production Zone (Maximum Security)

```terraform
resource "bind9_zone" "production" {
  name = "prod.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "security.example.com"
  soa_refresh = 86400
  soa_retry   = 7200
  soa_expire  = 3600000
  soa_minimum = 3600
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Internal only
  allow_query = ["internal"]
  
  # TSIG-authenticated transfers only
  allow_transfer = ["key transfer-key"]
  
  # No dynamic updates in production
  allow_update = ["none"]
  
  notify = true
  
  # Don't delete zone file on destroy
  delete_file_on_destroy = false
}
```

### Subdomain Delegation Zone

```terraform
resource "bind9_zone" "subdomain" {
  name = "sub.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Inherit from parent zone ACL
  allow_query = ["internal"]
  
  # Allow parent zone servers to transfer
  allow_transfer = [
    "trusted-secondaries",
    "10.0.1.1",  # Parent zone NS
  ]
  
  allow_update = ["key ddns-key"]
  
  notify = true
}
```

## Reverse DNS Zone ACLs

```terraform
resource "bind9_zone" "reverse" {
  name = "1.0.10.in-addr.arpa"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # PTR lookups from internal only
  allow_query = ["internal"]
  
  # Standard secondary transfers
  allow_transfer = ["trusted-secondaries"]
  
  # Allow DHCP server to create PTR records
  allow_update = [
    "10.0.2.5",    # DHCP server IP
    "key ddns-key",
  ]
  
  notify = true
}
```

## TSIG Key Setup

To use TSIG key authentication, you need to:

### 1. Generate TSIG Key on BIND9 Server

```bash
# Generate key
tsig-keygen -a hmac-sha256 ddns-key > /etc/bind/keys/ddns-key.key
tsig-keygen -a hmac-sha256 transfer-key > /etc/bind/keys/transfer-key.key

# Set permissions
chown bind:bind /etc/bind/keys/*.key
chmod 640 /etc/bind/keys/*.key
```

### 2. Include Keys in named.conf

```bind
include "/etc/bind/keys/ddns-key.key";
include "/etc/bind/keys/transfer-key.key";
```

### 3. Reference in Terraform

```terraform
allow_update   = ["key ddns-key"]
allow_transfer = ["key transfer-key"]
```

## Best Practices

### Security Recommendations

| Setting | Recommendation |
|---------|----------------|
| `allow_query` | Use `internal` ACL for private zones |
| `allow_transfer` | Always use TSIG keys or specific IPs |
| `allow_update` | Use TSIG keys, never `any` |
| Production zones | Set `allow_update = ["none"]` |

### Common Patterns

| Zone Type | allow_query | allow_transfer | allow_update |
|-----------|-------------|----------------|--------------|
| Public | `any` | `trusted-secondaries` | `none` |
| Internal | `internal` | `trusted-secondaries` | `key ddns-key` |
| DMZ | `internal`, `dmz` | `localnets` | `key ddns-key` |
| Production | `internal` | `key transfer-key` | `none` |
| Reverse DNS | `internal` | `trusted-secondaries` | `key ddns-key` |

### Troubleshooting

| Issue | Possible Cause | Solution |
|-------|---------------|----------|
| Query refused | Client IP not in `allow_query` | Add client network to ACL |
| Transfer failed | Server not in `allow_transfer` | Add secondary NS IP or use TSIG |
| Update refused | Client not in `allow_update` | Add client IP or use TSIG key |
| Named ACL not found | ACL not defined in named.conf | Define ACL before zone |

## Complete Example

Here's a complete example with multiple zones using different ACL patterns:

```terraform
# Public zone
resource "bind9_zone" "public" {
  name        = "example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
  allow_query    = ["any"]
  allow_transfer = ["trusted-secondaries"]
  allow_update   = ["none"]
  notify         = true
}

# Internal zone
resource "bind9_zone" "internal" {
  name        = "internal.example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
  allow_query    = ["internal"]
  allow_transfer = ["trusted-secondaries"]
  allow_update   = ["ddns-clients"]
  notify         = true
}

# Reverse zone
resource "bind9_zone" "reverse" {
  name        = "1.0.10.in-addr.arpa"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
  allow_query    = ["internal"]
  allow_transfer = ["trusted-secondaries"]
  allow_update   = ["ddns-clients"]
  notify         = true
}
```
