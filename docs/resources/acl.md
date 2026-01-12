---
page_title: "bind9_acl Resource - BIND9 Provider"
subcategory: "Resources"
description: |-
  Manages a BIND9 Access Control List (ACL).
---

# bind9_acl (Resource)

Manages a BIND9 Access Control List (ACL). ACLs are named groups of IP addresses, networks, or TSIG keys that can be referenced in zone configurations for access control.

## Why Use ACLs?

Instead of repeating the same list of IP addresses in every zone, you can define an ACL once and reference it by name:

```terraform
# Without ACL - repetitive and error-prone
resource "bind9_zone" "zone1" {
  allow_query    = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
  allow_transfer = ["10.0.1.11", "10.0.1.12"]
}

resource "bind9_zone" "zone2" {
  allow_query    = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
  allow_transfer = ["10.0.1.11", "10.0.1.12"]
}

# With ACL - clean and maintainable
resource "bind9_acl" "internal" {
  name    = "internal"
  entries = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}

resource "bind9_acl" "secondaries" {
  name    = "secondaries"
  entries = ["10.0.1.11", "10.0.1.12"]
}

resource "bind9_zone" "zone1" {
  allow_query    = ["internal"]
  allow_transfer = ["secondaries"]
  depends_on     = [bind9_acl.internal, bind9_acl.secondaries]
}
```

## Example Usage

### Basic ACL - Internal Networks

```terraform
resource "bind9_acl" "internal" {
  name = "internal"
  
  entries = [
    "localhost",
    "localnets",
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
  ]
  
  comment = "RFC1918 private networks"
}
```

### ACL with TSIG Key Authentication

```terraform
resource "bind9_acl" "ddns_clients" {
  name = "ddns-clients"
  
  entries = [
    "172.25.44.0/24",
    "key \"ddns-key\"",     # TSIG key authentication
  ]
  
  comment = "Clients allowed for dynamic DNS updates"
}
```

### ACL for Zone Transfers

```terraform
resource "bind9_acl" "trusted_secondaries" {
  name = "trusted-secondaries"
  
  entries = [
    "172.25.44.78",          # Secondary NS 1
    "172.25.44.79",          # Secondary NS 2
    "key \"transfer-key\"",  # TSIG authenticated transfers
  ]
  
  comment = "Secondary DNS servers allowed for zone transfers"
}
```

### ACL for External Access

```terraform
resource "bind9_acl" "external_trusted" {
  name = "external-trusted"
  
  entries = [
    "8.8.8.8",               # Google DNS
    "1.1.1.1",               # Cloudflare DNS
    "203.0.113.0/24",        # Partner network
  ]
  
  comment = "External trusted networks"
}
```

### Complete Example - Using ACLs with Zones

```terraform
# =============================================================================
# Define ACLs
# =============================================================================

resource "bind9_acl" "internal" {
  name    = "internal"
  entries = ["localhost", "localnets", "10.0.0.0/8"]
  comment = "Internal networks"
}

resource "bind9_acl" "secondaries" {
  name    = "secondaries"
  entries = ["10.0.1.11", "10.0.1.12", "key \"transfer-key\""]
  comment = "Secondary DNS servers"
}

resource "bind9_acl" "ddns" {
  name    = "ddns"
  entries = ["10.0.2.0/24", "key \"ddns-key\""]
  comment = "Dynamic DNS update clients"
}

# =============================================================================
# Use ACLs in Zone Configuration
# =============================================================================

resource "bind9_zone" "example" {
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

  # Reference ACLs by name
  allow_query    = ["internal", "any"]     # Internal + public queries
  allow_transfer = ["secondaries"]         # Only to secondaries
  allow_update   = ["ddns"]                # Only DDNS clients
  
  notify = true

  # Ensure ACLs are created before the zone
  depends_on = [
    bind9_acl.internal,
    bind9_acl.secondaries,
    bind9_acl.ddns,
  ]
}
```

### Multi-Server ACL Deployment

Deploy the same ACLs to multiple BIND9 servers:

```terraform
# =============================================================================
# ACLs for dns1
# =============================================================================

resource "bind9_acl" "internal_dns1" {
  count    = try(var.servers["dns1"].enabled, false) ? 1 : 0
  provider = bind9.dns1

  name    = "internal"
  entries = ["localhost", "localnets", "10.0.0.0/8", "172.16.0.0/12"]
  comment = "Internal networks"
}

resource "bind9_acl" "secondaries_dns1" {
  count    = try(var.servers["dns1"].enabled, false) ? 1 : 0
  provider = bind9.dns1

  name    = "secondaries"
  entries = ["10.0.1.11", "10.0.1.12"]
  comment = "Secondary DNS servers"
}

# =============================================================================
# ACLs for dns2
# =============================================================================

resource "bind9_acl" "internal_dns2" {
  count    = try(var.servers["dns2"].enabled, false) ? 1 : 0
  provider = bind9.dns2

  name    = "internal"
  entries = ["localhost", "localnets", "10.0.0.0/8", "172.16.0.0/12"]
  comment = "Internal networks"
}

resource "bind9_acl" "secondaries_dns2" {
  count    = try(var.servers["dns2"].enabled, false) ? 1 : 0
  provider = bind9.dns2

  name    = "secondaries"
  entries = ["10.0.1.10", "10.0.1.12"]  # Different for dns2
  comment = "Secondary DNS servers"
}
```

## Schema

### Required

- `name` (String) Name of the ACL. Must be a valid identifier (letters, numbers, underscores, hyphens).
- `entries` (List of String) List of ACL entries.

### Optional

- `comment` (String) Description or comment for the ACL.

### Read-Only

- `id` (String) ACL identifier (same as name).

## ACL Entry Formats

The `entries` attribute accepts various formats:

| Format | Example | Description |
|--------|---------|-------------|
| `localhost` | `"localhost"` | Local machine |
| `localnets` | `"localnets"` | All local networks |
| IP address | `"192.168.1.5"` | Single host |
| CIDR network | `"10.0.0.0/8"` | Network range |
| TSIG key | `"key \"keyname\""` | TSIG key authentication |
| `any` | `"any"` | Any host |
| `none` | `"none"` | No hosts |
| Nested ACL | `"internal"` | Reference another ACL by name |

### TSIG Key Format

To reference a TSIG key in an ACL:

```terraform
entries = [
  "key \"ddns-key\"",      # Note the escaped quotes
]
```

The key must be defined in BIND9's configuration (`/etc/bind/named.conf` or included files).

## Import

Import an existing ACL by name:

```bash
terraform import bind9_acl.internal internal
```

## BIND9 Server Requirements

For ACL management to work, your BIND9 server needs specific configuration.

### Required Configuration Files

| File | Purpose | Permissions |
|------|---------|-------------|
| `/etc/bind/named.conf` | Main BIND9 config | `root:bind` `644` |
| `/etc/bind/named.conf.acls` | ACL definitions (managed by API) | `bind:bind` `664` |
| `/etc/bind/rndc.key` | RNDC authentication | `bind:bind` `640` |
| `/etc/bind/keys/ddns-key.key` | TSIG key for updates | `bind:bind` `640` |

### Step 1: Create ACL Configuration File

The API manages ACLs in `/etc/bind/named.conf.acls`. Create this file:

```bash
# Create the file
touch /etc/bind/named.conf.acls

# Set permissions so API can write to it
chown bind:bind /etc/bind/named.conf.acls
chmod 664 /etc/bind/named.conf.acls
```

### Step 2: Include ACL File in BIND9 Configuration

Add this include statement to `/etc/bind/named.conf` **BEFORE** any zone definitions:

```bind
// /etc/bind/named.conf

// Include keys first
include "/etc/bind/rndc.key";
include "/etc/bind/keys/ddns-key.key";

// Include API-managed ACLs - MUST be before zones!
include "/etc/bind/named.conf.acls";

// Then include zones
include "/etc/bind/named.conf.options";
include "/etc/bind/named.conf.local";
include "/etc/bind/named.conf.default-zones";
```

### Step 3: Configure API Environment

Add the ACL file path to the API's `.env` configuration:

```bash
# /opt/bind9-api/.env
BIND9_API_BIND9_ACLS_PATH=/etc/bind/named.conf.acls
```

### Step 4: Reload BIND9

```bash
# Check configuration is valid
named-checkconf

# Reload BIND9 to pick up the include
rndc reconfig
```

### Verification

After creating an ACL via Terraform, verify it's in the file:

```bash
# View managed ACLs
cat /etc/bind/named.conf.acls

# Verify BIND9 loaded it
rndc status
```

## How ACLs Are Stored

The provider manages ACLs in `/etc/bind/named.conf.acls`. This file is:
- Created automatically if it doesn't exist
- Managed entirely by the API - **DO NOT EDIT MANUALLY**
- Included in BIND9 config via `include "/etc/bind/named.conf.acls";`

Example generated file:

```bind
// BIND9 Access Control Lists
// Managed by bind9-api - DO NOT EDIT MANUALLY
// Last updated: 2026-01-12T10:30:00

// RFC1918 private networks
acl "internal" {
    localhost;
    localnets;
    10.0.0.0/8;
    172.16.0.0/12;
    192.168.0.0/16;
};

// Dynamic DNS update clients
acl "ddns-clients" {
    172.25.44.0/24;
    key "ddns-key";
};

// Secondary DNS servers
acl "secondaries" {
    10.0.1.11;
    10.0.1.12;
    key "transfer-key";
};
```

## See Also

- [Access Control Guide](../guides/access-control.md) - Detailed guide on BIND9 access control
- [bind9_zone Resource](zone.md) - Zone resource that uses ACLs
- [Multi-Server Guide](../guides/multi-server.md) - Deploying ACLs to multiple servers
