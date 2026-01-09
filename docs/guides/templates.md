---
page_title: "Ready-to-Use Templates"
subcategory: "Guides"
description: |-
  Complete, ready-to-use Terraform configuration templates for common DNS scenarios.
---

# Ready-to-Use Templates

This guide provides complete, copy-paste ready Terraform configurations for common DNS scenarios. Just replace the placeholder values with your own.

## Template 1: Single Zone with Common Records

A complete setup for a single domain with web, mail, and common DNS records.

```terraform
# =============================================================================
# providers.tf - Provider Configuration
# =============================================================================

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

provider "bind9" {
  endpoint = var.bind9_endpoint  # e.g., "https://dns.example.com:8080"
  api_key  = var.bind9_api_key
}

# =============================================================================
# variables.tf - Input Variables
# =============================================================================

variable "bind9_endpoint" {
  description = "BIND9 REST API endpoint URL"
  type        = string
}

variable "bind9_api_key" {
  description = "BIND9 REST API key"
  type        = string
  sensitive   = true
}

variable "domain" {
  description = "Domain name to manage"
  type        = string
  default     = "example.com"
}

variable "ns1_ip" {
  description = "Primary nameserver IP"
  type        = string
  default     = "10.0.1.10"
}

variable "ns2_ip" {
  description = "Secondary nameserver IP"
  type        = string
  default     = "10.0.1.11"
}

variable "web_server_ip" {
  description = "Web server IP address"
  type        = string
  default     = "10.0.1.100"
}

variable "mail_server_ip" {
  description = "Mail server IP address"
  type        = string
  default     = "10.0.1.50"
}

# =============================================================================
# zones.tf - Zone Configuration
# =============================================================================

resource "bind9_zone" "main" {
  name = var.domain
  type = "master"

  soa_mname   = "ns1.${var.domain}"
  soa_rname   = "hostmaster.${var.domain}"
  soa_refresh = 86400
  soa_retry   = 7200
  soa_expire  = 3600000
  soa_minimum = 3600
  default_ttl = 3600

  nameservers = [
    "ns1.${var.domain}",
    "ns2.${var.domain}",
  ]

  ns_addresses = {
    "ns1.${var.domain}" = var.ns1_ip
    "ns2.${var.domain}" = var.ns2_ip
  }

  allow_query    = ["any"]
  allow_transfer = [var.ns2_ip]
  allow_update   = ["none"]
  
  notify = true
}

# =============================================================================
# records.tf - DNS Records
# =============================================================================

# Zone apex A record
resource "bind9_record" "apex" {
  zone    = bind9_zone.main.name
  name    = "@"
  type    = "A"
  ttl     = 300
  records = [var.web_server_ip]
}

# www subdomain
resource "bind9_record" "www" {
  zone    = bind9_zone.main.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = [var.web_server_ip]
}

# Mail server
resource "bind9_record" "mail" {
  zone    = bind9_zone.main.name
  name    = "mail"
  type    = "A"
  ttl     = 300
  records = [var.mail_server_ip]
}

# MX record
resource "bind9_record" "mx" {
  zone    = bind9_zone.main.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.${var.domain}."]
}

# SPF record
resource "bind9_record" "spf" {
  zone    = bind9_zone.main.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 mx -all"]
}

# DMARC record
resource "bind9_record" "dmarc" {
  zone    = bind9_zone.main.name
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DMARC1; p=reject; rua=mailto:dmarc@${var.domain}"]
}

# =============================================================================
# outputs.tf - Output Values
# =============================================================================

output "zone_name" {
  value = bind9_zone.main.name
}

output "zone_serial" {
  value = bind9_zone.main.serial
}

output "nameservers" {
  value = bind9_zone.main.nameservers
}
```

---

## Template 2: Multi-Server Deployment

Deploy zones and records to multiple BIND9 servers simultaneously.

```terraform
# =============================================================================
# providers.tf - Multi-Server Provider Configuration
# =============================================================================

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# =============================================================================
# variables.tf
# =============================================================================

variable "servers" {
  description = "Map of BIND9 servers"
  type = map(object({
    endpoint = string
    api_key  = string
    enabled  = bool
  }))
}

# =============================================================================
# terraform.tfvars (create separately, add to .gitignore)
# =============================================================================
# servers = {
#   "dns1" = {
#     endpoint = "https://dns1.example.com:8080"
#     api_key  = "your-api-key-1"
#     enabled  = true
#   }
#   "dns2" = {
#     endpoint = "https://dns2.example.com:8080"
#     api_key  = "your-api-key-2"
#     enabled  = true
#   }
# }

# =============================================================================
# providers.tf - Server Aliases
# =============================================================================

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

# Default provider
provider "bind9" {
  endpoint = var.servers["dns1"].endpoint
  api_key  = var.servers["dns1"].api_key
}

# =============================================================================
# locals.tf - Record Definitions
# =============================================================================

locals {
  enabled_servers = {
    for name, server in var.servers : name => server
    if server.enabled
  }

  # Define records once, deploy to all servers
  records = {
    "www_A" = {
      name    = "www"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.100"]
      servers = []  # Empty = all servers
    }
    "api_A" = {
      name    = "api"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.101"]
      servers = []
    }
    "app_A" = {
      name    = "app"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.102"]
      servers = []
    }
    "mail_A" = {
      name    = "mail"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.50"]
      servers = []
    }
    "mx_MX" = {
      name    = "@"
      type    = "MX"
      ttl     = 3600
      records = ["10 mail.example.com."]
      servers = []
    }
    "spf_TXT" = {
      name    = "@"
      type    = "TXT"
      ttl     = 3600
      records = ["v=spf1 mx -all"]
      servers = []
    }
  }

  # Expand records to target servers
  records_expanded = merge([
    for record_key, record in local.records : {
      for server_name, server in local.enabled_servers :
      "${record_key}_${server_name}" => merge(record, { server = server_name })
      if length(record.servers) == 0 || contains(record.servers, server_name)
    }
  ]...)
}

# =============================================================================
# zones.tf - Zones on Each Server
# =============================================================================

resource "bind9_zone" "main_dns1" {
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

  allow_query    = ["any"]
  allow_transfer = ["10.0.1.11"]
  notify         = true
}

resource "bind9_zone" "main_dns2" {
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

  allow_query    = ["any"]
  allow_transfer = ["10.0.1.10"]
  notify         = true
}

# =============================================================================
# records.tf - Records on Each Server
# =============================================================================

resource "bind9_record" "main_dns1" {
  for_each = {
    for k, v in local.records_expanded : k => v
    if v.server == "dns1"
  }
  provider = bind9.dns1

  zone    = bind9_zone.main_dns1[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}

resource "bind9_record" "main_dns2" {
  for_each = {
    for k, v in local.records_expanded : k => v
    if v.server == "dns2"
  }
  provider = bind9.dns2

  zone    = bind9_zone.main_dns2[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

---

## Template 3: Bulk Host Generation

Generate multiple hosts with sequential IPs (replaces BIND9 $GENERATE).

```terraform
# =============================================================================
# Bulk Host Generation Template
# =============================================================================

variable "host_prefix" {
  default = "host"
}

variable "host_count" {
  default = 50
}

variable "base_ip_prefix" {
  default = "10.0.2"
}

variable "start_octet" {
  default = 1
}

# Zone setup (abbreviated)
resource "bind9_zone" "hosts" {
  name        = "hosts.example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
}

# Generate A records: host-1, host-2, ..., host-50
resource "bind9_record" "hosts" {
  for_each = {
    for i in range(1, var.host_count + 1) : "${var.host_prefix}-${i}" => {
      name = "${var.host_prefix}-${i}"
      ip   = "${var.base_ip_prefix}.${var.start_octet + i - 1}"
    }
  }

  zone    = bind9_zone.hosts.name
  name    = each.value.name
  type    = "A"
  ttl     = 300
  records = [each.value.ip]
}

# Corresponding reverse zone
resource "bind9_zone" "reverse" {
  name        = "2.0.10.in-addr.arpa"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600
  nameservers = ["ns1.example.com", "ns2.example.com"]
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
}

# Generate PTR records
resource "bind9_record" "ptrs" {
  for_each = {
    for i in range(1, var.host_count + 1) : "${var.start_octet + i - 1}" => {
      octet = "${var.start_octet + i - 1}"
      fqdn  = "${var.host_prefix}-${i}.hosts.example.com."
    }
  }

  zone    = bind9_zone.reverse.name
  name    = each.value.octet
  type    = "PTR"
  ttl     = 300
  records = [each.value.fqdn]
}

# Output generated hosts
output "generated_hosts" {
  value = [
    for i in range(1, var.host_count + 1) : {
      hostname = "${var.host_prefix}-${i}.hosts.example.com"
      ip       = "${var.base_ip_prefix}.${var.start_octet + i - 1}"
    }
  ]
}
```

---

## Template 4: Complete Zone with DNSSEC

A zone with DNSSEC enabled.

```terraform
# =============================================================================
# DNSSEC-Enabled Zone Template
# =============================================================================

# Zone
resource "bind9_zone" "secure" {
  name = "secure.example.com"
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

  allow_query    = ["any"]
  allow_transfer = ["key transfer-key"]
  allow_update   = ["none"]
  
  notify                 = true
  delete_file_on_destroy = false
}

# KSK (Key Signing Key)
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.secure.name
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  ttl       = 3600
}

# ZSK (Zone Signing Key)
resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.secure.name
  key_type  = "ZSK"
  algorithm = 13
  ttl       = 300
  sign_zone = true  # Sign zone after creating ZSK
}

# Records
resource "bind9_record" "www" {
  zone    = bind9_zone.secure.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}

# Output DS records for registrar
output "ds_records" {
  value       = bind9_dnssec_key.ksk.ds_records
  description = "Submit these DS records to your domain registrar"
}

output "dnssec_status" {
  value = {
    zone           = bind9_zone.secure.name
    dnssec_enabled = bind9_zone.secure.dnssec_enabled
    ksk_key_tag    = bind9_dnssec_key.ksk.key_tag
    zsk_key_tag    = bind9_dnssec_key.zsk.key_tag
  }
}
```

---

## Template 5: Internal Network Zones

Complete internal DNS setup with forward and reverse zones.

```terraform
# =============================================================================
# Internal Network DNS Template
# =============================================================================

variable "internal_domain" {
  default = "internal.example.com"
}

variable "networks" {
  default = {
    "servers"    = { cidr = "10.0.1.0/24", reverse = "1.0.10.in-addr.arpa" }
    "workstations" = { cidr = "10.0.2.0/24", reverse = "2.0.10.in-addr.arpa" }
    "printers"   = { cidr = "10.0.3.0/24", reverse = "3.0.10.in-addr.arpa" }
  }
}

# Forward zone
resource "bind9_zone" "internal" {
  name = var.internal_domain
  type = "master"

  soa_mname   = "ns1.${var.internal_domain}"
  soa_rname   = "hostmaster.${var.internal_domain}"
  default_ttl = 3600

  nameservers = [
    "ns1.${var.internal_domain}",
    "ns2.${var.internal_domain}",
  ]
  ns_addresses = {
    "ns1.${var.internal_domain}" = "10.0.1.10"
    "ns2.${var.internal_domain}" = "10.0.1.11"
  }

  allow_query    = ["10.0.0.0/8"]
  allow_transfer = ["10.0.1.11"]
  allow_update   = ["key ddns-key"]
  
  notify = true
}

# Reverse zones
resource "bind9_zone" "reverse" {
  for_each = var.networks

  name = each.value.reverse
  type = "master"

  soa_mname   = "ns1.${var.internal_domain}"
  soa_rname   = "hostmaster.${var.internal_domain}"
  default_ttl = 3600

  nameservers = [
    "ns1.${var.internal_domain}",
    "ns2.${var.internal_domain}",
  ]
  ns_addresses = {
    "ns1.${var.internal_domain}" = "10.0.1.10"
    "ns2.${var.internal_domain}" = "10.0.1.11"
  }

  allow_query    = ["10.0.0.0/8"]
  allow_transfer = ["10.0.1.11"]
  allow_update   = ["key ddns-key"]
  
  notify = true
}

# Common internal services
locals {
  internal_services = {
    "ldap"     = "10.0.1.20"
    "nfs"      = "10.0.1.21"
    "gitlab"   = "10.0.1.22"
    "jenkins"  = "10.0.1.23"
    "registry" = "10.0.1.24"
    "vault"    = "10.0.1.25"
    "consul"   = "10.0.1.26"
  }
}

resource "bind9_record" "services" {
  for_each = local.internal_services

  zone    = bind9_zone.internal.name
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}

output "internal_services" {
  value = {
    for name, ip in local.internal_services :
    name => "${name}.${var.internal_domain}"
  }
}
```

---

## Quick Start Checklist

1. ✅ Copy the template that matches your use case
2. ✅ Create `terraform.tfvars` with your values
3. ✅ Add `terraform.tfvars` to `.gitignore`
4. ✅ Run `terraform init`
5. ✅ Run `terraform plan` to preview
6. ✅ Run `terraform apply` to create

## .gitignore Template

```gitignore
# Terraform
.terraform/
.terraform.lock.hcl
*.tfstate
*.tfstate.*
*.tfplan

# Sensitive files
terraform.tfvars
*.auto.tfvars
secrets.tf

# IDE
.idea/
.vscode/
*.swp
```
