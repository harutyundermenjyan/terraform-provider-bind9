---
page_title: "Getting Started with BIND9 Provider"
subcategory: "Guides"
description: |-
  A quick start guide for using the BIND9 Terraform provider.
---

# Getting Started

This guide walks you through setting up the BIND9 provider and creating your first DNS zone and records.

## Prerequisites

Before you begin, you need:

1. **BIND9 REST API** - A running instance of the [BIND9 REST API](https://github.com/harutyundermenjyan/bind9-api)
2. **API Key** - An API key or credentials for authentication
3. **Terraform/OpenTofu** - Version 1.0 or later

## Step 1: Configure the Provider

Create a new Terraform configuration file:

```terraform
# main.tf

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# Configure the BIND9 provider
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  api_key  = var.bind9_api_key
}

# Variables
variable "bind9_api_key" {
  type        = string
  sensitive   = true
  description = "API key for BIND9 REST API"
}
```

## Step 2: Create Your First Zone

Add a zone resource:

```terraform
# zones.tf

resource "bind9_zone" "myzone" {
  name = "myzone.example.com"
  type = "master"

  # SOA Record settings
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 3600      # 1 hour
  soa_retry   = 600       # 10 minutes
  soa_expire  = 604800    # 1 week
  soa_minimum = 86400     # 1 day

  # Default TTL for records
  default_ttl = 3600

  # Nameservers
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]

  # Glue records (IP addresses for in-zone nameservers)
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }

  # Access control
  allow_transfer = ["10.0.1.11"]  # Allow zone transfers to secondary
  allow_update   = ["key ddns-key"]  # Allow dynamic updates with TSIG key
  notify         = true  # Notify secondaries on changes
}
```

## Step 3: Create DNS Records

Add some records to your zone:

```terraform
# records.tf

# Web server
resource "bind9_record" "www" {
  zone    = bind9_zone.myzone.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}

# Web server alias
resource "bind9_record" "site" {
  zone    = bind9_zone.myzone.name
  name    = "site"
  type    = "CNAME"
  ttl     = 3600
  records = ["www.myzone.example.com."]  # Note: trailing dot for FQDN
}

# Mail server
resource "bind9_record" "mail" {
  zone    = bind9_zone.myzone.name
  name    = "mail"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.50"]
}

# MX records for email
resource "bind9_record" "mx" {
  zone    = bind9_zone.myzone.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = [
    "10 mail.myzone.example.com.",
  ]
}

# SPF record for email security
resource "bind9_record" "spf" {
  zone    = bind9_zone.myzone.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 mx -all"]
}
```

## Step 4: Initialize and Apply

```bash
# Set your API key
export TF_VAR_bind9_api_key="your-api-key-here"

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

## Step 5: Verify Your Configuration

Query your zone information:

```terraform
# data.tf

data "bind9_zone" "check" {
  name = bind9_zone.myzone.name
}

output "zone_serial" {
  value = data.bind9_zone.check.serial
}

output "zone_loaded" {
  value = data.bind9_zone.check.loaded
}

data "bind9_records" "all" {
  zone = bind9_zone.myzone.name
}

output "all_records" {
  value = data.bind9_records.all.records
}
```

## Common Patterns

### Multiple Servers

```terraform
provider "bind9" {
  alias    = "primary"
  endpoint = "https://dns1.example.com:8080"
  api_key  = var.primary_api_key
}

provider "bind9" {
  alias    = "secondary"
  endpoint = "https://dns2.example.com:8080"
  api_key  = var.secondary_api_key
}

resource "bind9_zone" "primary" {
  provider = bind9.primary
  name     = "example.com"
  type     = "master"
  # ...
}

resource "bind9_zone" "secondary" {
  provider = bind9.secondary
  name     = "example.com"
  type     = "slave"
}
```

### Define Records Once, Deploy to Multiple Servers

```terraform
locals {
  records = {
    "www_A" = {
      name    = "www"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.100"]
    }
    "app_A" = {
      name    = "app"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.101"]
    }
  }
}

resource "bind9_record" "primary" {
  for_each = local.records
  provider = bind9.primary

  zone    = bind9_zone.primary.name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}

resource "bind9_record" "secondary" {
  for_each = local.records
  provider = bind9.secondary

  zone    = bind9_zone.secondary.name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

### Generate Multiple Records (like BIND9 $GENERATE)

```terraform
# Create host-1 through host-10 with sequential IPs
resource "bind9_record" "hosts" {
  for_each = { for i in range(1, 11) : "host-${i}" => i }

  zone    = bind9_zone.myzone.name
  name    = each.key
  type    = "A"
  ttl     = 300
  records = ["10.0.1.${100 + each.value}"]
}
```

### Reverse DNS Zone

```terraform
# Reverse zone for 10.0.1.0/24
resource "bind9_zone" "reverse" {
  name = "1.0.10.in-addr.arpa"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  nameservers = ["ns1.example.com", "ns2.example.com"]
}

# PTR record for 10.0.1.100
resource "bind9_record" "ptr_100" {
  zone    = bind9_zone.reverse.name
  name    = "100"
  type    = "PTR"
  ttl     = 3600
  records = ["www.myzone.example.com."]
}
```

## Next Steps

- Read the [Zone Resource Documentation](../resources/zone.md) for all zone options
- Learn about [Record Types](../resources/record.md) and their formats
- Set up [DNSSEC](../resources/dnssec_key.md) for your zones
- Use [Data Sources](../data-sources/zone.md) to query existing configurations

## Troubleshooting

### Connection Errors

```
Error: Could not create BIND9 API client
```

- Verify the endpoint URL is correct
- Check that the API server is running
- Ensure network connectivity between Terraform and the API

### Authentication Errors

```
Error: API error 401: Invalid authentication credentials
```

- Verify your API key is correct
- Check that the API key has required permissions
- Ensure the API key hasn't expired

### Zone Validation Errors

```
Error: In-zone nameserver requires an IP address in ns_addresses
```

- Provide IP addresses for nameservers within the zone
- Use the `ns_addresses` map for glue records

### Record Format Errors

- Always use trailing dots for FQDNs in CNAME, MX, NS, PTR targets
- Format MX records as "priority hostname" (e.g., "10 mail.example.com.")
- Format SRV records as "priority weight port target"
