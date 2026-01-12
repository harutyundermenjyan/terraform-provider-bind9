---
page_title: "Multi-Server Setup Guide"
subcategory: "Guides"
description: |-
  Learn how to manage multiple BIND9 servers with a single Terraform/OpenTofu configuration.
---

# Multi-Server Setup Guide

This guide explains how to manage multiple BIND9 servers using provider aliases and reusable configuration patterns.

## Architecture Options

### Single Server

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   Terraform/    │──────▶│   BIND9 REST    │──────▶│     BIND9       │
│   OpenTofu      │       │      API        │       │     Server      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
```

### Multi-Primary Setup (Independent Primaries)

Each BIND9 server acts as an independent primary. Zones and records are deployed to all servers via Terraform.

```
                         ┌─────────────────────────────┐
                         │      Terraform/OpenTofu     │
                         │     (Single Configuration)  │
                         └──────────────┬──────────────┘
                                        │
              ┌─────────────────────────┼─────────────────────────┐
              │                         │                         │
              ▼                         ▼                         ▼
     ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
     │   BIND9 API     │       │   BIND9 API     │       │   BIND9 API     │
     │   (Server 1)    │       │   (Server 2)    │       │   (Server 3)    │
     └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
              │                         │                         │
              ▼                         ▼                         ▼
     ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
     │  BIND9 Primary  │       │  BIND9 Primary  │       │  BIND9 Primary  │
     │  (Identical)    │       │  (Identical)    │       │  (Identical)    │
     └─────────────────┘       └─────────────────┘       └─────────────────┘
```

### Primary/Secondary Setup (Traditional)

One server is the primary, others transfer zones via AXFR/IXFR.

```
     ┌─────────────────┐
     │   Terraform     │
     └────────┬────────┘
              │
              ▼
     ┌─────────────────┐       AXFR/IXFR      ┌─────────────────┐
     │  BIND9 Primary  │─────────────────────▶│  BIND9 Secondary│
     │   (master)      │                      │    (slave)      │
     └─────────────────┘                      └─────────────────┘
```

## Provider Configuration

### Step 1: Define Variables

```terraform
# variables.tf
variable "servers" {
  description = "Map of BIND9 servers"
  type = map(object({
    endpoint = string
    api_key  = string
    enabled  = bool
  }))
}
```

### Step 2: Configure Server Values

```terraform
# terraform.tfvars
servers = {
  "dns1" = {
    endpoint = "https://dns1.example.com:8080"
    api_key  = "your-api-key-for-dns1"
    enabled  = true
  }
  "dns2" = {
    endpoint = "https://dns2.example.com:8080"
    api_key  = "your-api-key-for-dns2"
    enabled  = true
  }
  "dns3" = {
    endpoint = "https://dns3.example.com:8080"
    api_key  = "your-api-key-for-dns3"
    enabled  = false  # Disabled temporarily
  }
}
```

### Step 3: Configure Provider Aliases

```terraform
# providers.tf
terraform {
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# Provider for each server
provider "bind9" {
  alias    = "dns1"
  endpoint = var.servers["dns1"].endpoint
  api_key  = var.servers["dns1"].api_key
}

provider "bind9" {
  alias    = "dns2"
  endpoint = var.servers["dns2"].endpoint
  api_key  = var.servers["dns2"].api_key
}

provider "bind9" {
  alias    = "dns3"
  endpoint = var.servers["dns3"].endpoint
  api_key  = var.servers["dns3"].api_key
}

# Default provider (required by OpenTofu)
provider "bind9" {
  endpoint = var.servers["dns1"].endpoint
  api_key  = var.servers["dns1"].api_key
}
```

## Zone Configuration

### Option 1: Individual Zone Resources

Create separate resources for each server:

```terraform
# zones.tf
resource "bind9_zone" "example_dns1" {
  count    = var.servers["dns1"].enabled ? 1 : 0
  provider = bind9.dns1

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
}

resource "bind9_zone" "example_dns2" {
  count    = var.servers["dns2"].enabled ? 1 : 0
  provider = bind9.dns2

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
}
```

### Option 2: Primary/Secondary Configuration

```terraform
# Primary zone on dns1
resource "bind9_zone" "example_primary" {
  provider = bind9.dns1

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

  # Allow zone transfer to secondary
  allow_transfer = ["10.0.1.11", "key transfer-key"]
  notify = true
}

# Secondary zone on dns2
resource "bind9_zone" "example_secondary" {
  provider = bind9.dns2

  name = "example.com"
  type = "slave"

  # Note: Master server is configured in named.conf on dns2
}
```

## Record Configuration with Multi-Server Deployment

### The `servers` Pattern

Define records once, deploy to selected servers using the `servers` attribute:

| `servers` Value | Behavior |
|-----------------|----------|
| `[]` (empty) | Deploy to **ALL** enabled servers |
| `["dns1"]` | Deploy to **dns1 only** |
| `["dns1", "dns2"]` | Deploy to **dns1 and dns2** |

```terraform
# locals.tf
locals {
  # Get only enabled servers
  enabled_servers = {
    for name, server in var.servers : name => server
    if server.enabled
  }

  # Define records with server targeting
  example_records = {
    # Public records → ALL servers
    "www_A" = {
      name    = "www"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.100"]
      servers = []              # ALL enabled servers
    }
    "mail_A" = {
      name    = "mail"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.50"]
      servers = []              # ALL enabled servers
    }
    "mx_MX" = {
      name    = "@"
      type    = "MX"
      ttl     = 3600
      records = ["10 mail.example.com."]
      servers = []              # ALL enabled servers
    }
    
    # Internal records → dns1 only
    "db_A" = {
      name    = "db"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.102"]
      servers = ["dns1"]        # dns1 only
    }
    "ldap_A" = {
      name    = "ldap"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.103"]
      servers = ["dns1"]        # dns1 only
    }
    
    # Staging records → dns2 only
    "staging_A" = {
      name    = "staging"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.200"]
      servers = ["dns2"]        # dns2 only
    }
    
    # Specific servers → dns1 and dns3
    "api_A" = {
      name    = "api"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.101"]
      servers = ["dns1", "dns3"]  # dns1 and dns3 only
    }
  }

  # Expand records to target servers
  example_records_expanded = merge([
    for record_key, record in local.example_records : {
      for server_name, server in local.enabled_servers :
      "${record_key}_${server_name}" => merge(record, { server = server_name })
      if length(record.servers) == 0 || contains(record.servers, server_name)
    }
  ]...)
}
```

### Create Records on Each Server

```terraform
# records.tf

# Records on dns1
resource "bind9_record" "example_dns1" {
  for_each = {
    for k, v in local.example_records_expanded : k => v
    if v.server == "dns1"
  }
  provider = bind9.dns1

  zone    = bind9_zone.example_dns1[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}

# Records on dns2
resource "bind9_record" "example_dns2" {
  for_each = {
    for k, v in local.example_records_expanded : k => v
    if v.server == "dns2"
  }
  provider = bind9.dns2

  zone    = bind9_zone.example_dns2[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

## Bulk Record Generation

Generate multiple records programmatically (equivalent to BIND9's `$GENERATE`):

```terraform
locals {
  # Generate host-1 through host-50 A records
  generated_hosts = {
    for i in range(1, 51) : "host-${i}_A" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
      servers = []  # All servers
    }
  }

  # Generate PTR records for reverse DNS
  generated_ptrs = {
    for i in range(1, 51) : "${i}_PTR" => {
      name    = "${i}"
      type    = "PTR"
      ttl     = 300
      records = ["host-${i}.example.com."]
      servers = []
    }
  }

  # Combine all records
  all_records = merge(
    local.example_records,
    local.generated_hosts,
    local.generated_ptrs
  )
}
```

## ACL Management on Multiple Servers

Deploy ACLs to each server to enable zone access control references.

### ACL Configuration per Server

```terraform
# =============================================================================
# acls.tf - ACLs for each server
# =============================================================================

# ACLs for dns1
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
  entries = ["10.0.1.11", "10.0.1.12", "key \"transfer-key\""]
  comment = "Secondary DNS servers"
}

resource "bind9_acl" "ddns_dns1" {
  count    = try(var.servers["dns1"].enabled, false) ? 1 : 0
  provider = bind9.dns1

  name    = "ddns-clients"
  entries = ["10.0.2.0/24", "key \"ddns-key\""]
  comment = "Dynamic DNS clients"
}

# ACLs for dns2
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
  entries = ["10.0.1.10", "10.0.1.12", "key \"transfer-key\""]  # Different for dns2
  comment = "Secondary DNS servers"
}

resource "bind9_acl" "ddns_dns2" {
  count    = try(var.servers["dns2"].enabled, false) ? 1 : 0
  provider = bind9.dns2

  name    = "ddns-clients"
  entries = ["10.0.2.0/24", "key \"ddns-key\""]
  comment = "Dynamic DNS clients"
}
```

### Using ACLs in Zone Configuration

```terraform
# zones.tf
resource "bind9_zone" "example_dns1" {
  count    = try(var.servers["dns1"].enabled, false) ? 1 : 0
  provider = bind9.dns1

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

  # Reference ACLs by name
  allow_query    = ["internal", "any"]
  allow_transfer = ["secondaries"]
  allow_update   = ["ddns-clients"]

  # Ensure ACLs are created first
  depends_on = [
    bind9_acl.internal_dns1,
    bind9_acl.secondaries_dns1,
    bind9_acl.ddns_dns1,
  ]
}
```

## Complete Example: File Structure

```
project/
├── providers.tf      # Provider configuration with aliases
├── variables.tf      # Server variable definitions
├── terraform.tfvars  # Server endpoints and API keys (gitignored)
├── acls.tf           # ACL resources for each server
├── locals.tf         # Record definitions and expansion logic
├── zones.tf          # Zone resources for each server
├── records.tf        # Record resources using for_each
└── outputs.tf        # Outputs for deployed resources
```

## Best Practices

### 1. Use Variables for Server Configuration

```terraform
# Keep credentials in terraform.tfvars (gitignored)
servers = {
  "dns1" = {
    endpoint = "https://dns1.example.com:8080"
    api_key  = "your-api-key"
    enabled  = true
  }
}
```

### 2. Enable/Disable Servers Without Removing Config

```terraform
# Just set enabled = false to temporarily disable
"dns3" = {
  endpoint = "https://dns3.example.com:8080"
  api_key  = "your-api-key"
  enabled  = false  # Won't create resources
}
```

### 3. Use Conditional Resource Creation

```terraform
resource "bind9_zone" "example_dns1" {
  count    = var.servers["dns1"].enabled ? 1 : 0
  provider = bind9.dns1
  # ...
}
```

### 4. Target Records to Specific Servers

```terraform
# Deploy to specific servers only
"internal_A" = {
  servers = ["dns1"]  # Internal record only on dns1
}

# Deploy to all servers
"www_A" = {
  servers = []  # Empty = all enabled servers
}
```

### 5. Separate Zones by File

For large deployments, organize zones into separate files:

```
zones/
├── example-com.tf
├── internal-example-com.tf
└── reverse-dns.tf
```

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Provider not found | Add the provider alias to `terraform.tfvars` |
| API connection error | Check endpoint URL and network access |
| Auth error (401) | Verify API key in terraform.tfvars |
| Zone exists error | Import existing zone: `terraform import bind9_zone.example example.com` |

### Verify Connectivity

```bash
# Test API access for each server
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://dns1.example.com:8080/api/v1/health

curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://dns2.example.com:8080/api/v1/health
```
