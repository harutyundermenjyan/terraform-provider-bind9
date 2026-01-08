---
page_title: "Getting Started with BIND9 Provider"
subcategory: "Guides"
description: |-
  Step-by-step guide to set up and use the BIND9 Terraform provider.
---

# Getting Started Guide

This guide walks you through setting up and using the BIND9 Terraform provider to manage your DNS infrastructure.

## Prerequisites

Before you begin, ensure you have:

1. **Terraform** >= 1.0 or **OpenTofu** >= 1.0 installed
2. **BIND9 REST API** ([bind9-api](https://gitlab.com/Dermenjyan/bind9-api)) deployed and running
3. **API credentials** (API key recommended)

## Step 1: Configure the Provider

Create a new Terraform configuration:

```hcl
# versions.tf
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# provider.tf
provider "bind9" {
  endpoint = var.bind9_endpoint
  api_key  = var.bind9_api_key
}

# variables.tf
variable "bind9_endpoint" {
  type        = string
  description = "BIND9 REST API endpoint"
}

variable "bind9_api_key" {
  type        = string
  description = "API key for authentication"
  sensitive   = true
}
```

Create a `terraform.tfvars` file (don't commit this to version control):

```hcl
bind9_endpoint = "http://dns.example.com:8080"
bind9_api_key  = "your-api-key-here"
```

## Step 2: Create Your First Zone

```hcl
# main.tf
resource "bind9_zone" "example" {
  name = "myzone.example.com"
  type = "master"
  
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
  
  allow_update = ["key ddns-key"]
}
```

Initialize and apply:

```bash
terraform init
terraform plan
terraform apply
```

## Step 3: Add DNS Records

### Basic A Record

```hcl
resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.2.100"]
}
```

### Multiple Records

```hcl
# MX records for email
resource "bind9_record" "mx" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = [
    "10 mail1.example.com.",
    "20 mail2.example.com.",
  ]
}

# TXT record for SPF
resource "bind9_record" "spf" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 mx ~all"]
}

# CNAME record
resource "bind9_record" "blog" {
  zone    = bind9_zone.example.name
  name    = "blog"
  type    = "CNAME"
  ttl     = 300
  records = ["www.myzone.example.com."]
}
```

## Step 4: Enable DNSSEC (Optional)

```hcl
# Key Signing Key
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.example.name
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}

# Zone Signing Key
resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.example.name
  key_type  = "ZSK"
  algorithm = 13
  
  depends_on = [bind9_dnssec_key.ksk]
}

# Output DS records for registrar
output "ds_records" {
  value     = bind9_dnssec_key.ksk.ds_records
  sensitive = false
}
```

## Step 5: Import Existing Resources

If you have existing zones and records, you can import them into Terraform.

### Import a Zone

```bash
terraform import bind9_zone.existing existing.example.com
```

### Import a Record

```bash
# Format: terraform import bind9_record.<name> <zone>/<record_name>/<type>
terraform import bind9_record.www example.com/www/A
```

## Common Patterns

### Environment-Specific Records

```hcl
variable "environment" {
  type = string
}

variable "server_ips" {
  type = map(list(string))
  default = {
    dev  = ["10.0.1.10"]
    prod = ["10.0.1.20", "10.0.1.21"]
  }
}

resource "bind9_record" "app" {
  zone    = bind9_zone.example.name
  name    = "app"
  type    = "A"
  ttl     = var.environment == "prod" ? 300 : 60
  records = var.server_ips[var.environment]
}
```

### Using Data Sources

```hcl
# Query existing records
data "bind9_record" "existing" {
  zone = "example.com"
  name = "legacy-server"
  type = "A"
}

# Create alias to existing server
resource "bind9_record" "alias" {
  zone    = "example.com"
  name    = "new-alias"
  type    = "A"
  ttl     = 300
  records = data.bind9_record.existing.records
}
```

### Bulk Record Generation

```hcl
# Generate multiple host records (like BIND9 $GENERATE)
locals {
  hosts = {
    for i in range(1, 11) :
    "host-${i}" => "10.0.3.${i}"
  }
}

resource "bind9_record" "hosts" {
  for_each = local.hosts
  
  zone    = bind9_zone.example.name
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}
```

## Troubleshooting

### Common Errors

**Error: "zone not found"**
- Ensure the zone exists or create it with `bind9_zone`
- Check that the zone name is correct

**Error: "authentication failed"**
- Verify your API key or credentials
- Check that the endpoint URL is correct

**Error: "connection refused"**
- Ensure the BIND9 REST API is running
- Check firewall rules

**Error: "NS has no address records"**
- Add `ns_addresses` to your zone definition for in-zone nameservers

### Enable Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply
```

## Next Steps

- Read the [Zone Resource](../resources/zone.md) documentation
- Read the [Record Resource](../resources/record.md) documentation
- Explore [DNSSEC Key Management](../resources/dnssec_key.md)
- Check out the [example configurations](https://github.com/harutyundermenjyan/terraform-provider-bind9/tree/main/examples)
