# Terraform Provider for BIND9

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-purple.svg)](https://registry.terraform.io/providers/harutyundermenjyan/bind9/latest)

A Terraform/OpenTofu provider for managing DNS zones and records on BIND9 servers via REST API.

---

## ⚠️ Required: BIND9 REST API

> **This provider requires the [BIND9 REST API](https://github.com/harutyundermenjyan/bind9-api) to be installed and running on your BIND9 server(s).**
>
> The provider communicates with BIND9 through this REST API - it does NOT connect directly to BIND9.
>
> 📦 **Get the API:** [github.com/harutyundermenjyan/bind9-api](https://github.com/harutyundermenjyan/bind9-api)

---

## Architecture Overview

This provider supports multiple deployment architectures. Choose the one that fits your needs:

### Architecture 1: Single Server

The simplest setup - one BIND9 server with one API instance.

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   Terraform/    │──────▶│   BIND9 REST    │──────▶│     BIND9       │
│   OpenTofu      │ HTTP  │      API        │ rndc  │     Server      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
     (this provider)           (:8080)              (DNS server)
```

**Use case:** Development, small deployments, single DNS server.

➡️ [Jump to Single Server Setup](#single-server-setup)

---

### Architecture 2: Multi-Primary Servers

Multiple independent BIND9 servers, each with its own API. Define records once, deploy to all or selected servers.

```
                          ┌─────────────────┐       ┌─────────────────┐
                     ┌───▶│   BIND9 API     │──────▶│     BIND9       │
                     │    │   (dns1:8080)   │       │   Server 1      │
                     │    └─────────────────┘       └─────────────────┘
┌─────────────────┐  │
│   Terraform/    │──┤    ┌─────────────────┐       ┌─────────────────┐
│   OpenTofu      │  ├───▶│   BIND9 API     │──────▶│     BIND9       │
│                 │  │    │   (dns2:8080)   │       │   Server 2      │
└─────────────────┘  │    └─────────────────┘       └─────────────────┘
                     │
                     │    ┌─────────────────┐       ┌─────────────────┐
                     └───▶│   BIND9 API     │──────▶│     BIND9       │
                          │   (dns3:8080)   │       │   Server 3      │
                          └─────────────────┘       └─────────────────┘
```

**Use case:** High availability, geographic distribution, separate environments.

**Key features:**
- Provider aliases for each server (`bind9.dns1`, `bind9.dns2`, etc.)
- `servers = []` pattern - define records once, deploy to all or selected servers
- `count` for conditional zone/record creation per server
- Independent servers (no replication between them)

➡️ [Jump to Multi-Server Setup](#multi-server-setup)

---

### Architecture Comparison

| Feature | Single Server | Multi-Primary |
|---------|---------------|---------------|
| BIND9 servers | 1 | 2+ |
| API instances | 1 | 1 per server |
| Provider blocks | 1 | 1 per server (aliases) |
| Record definition | Once | Once (with `servers = []` targeting) |
| Zone replication | N/A | Manual (same config) |
| Use case | Simple/Dev | HA/Geo/Multi-env |
| Complexity | Low | Medium |

---

## Features

- **Zone Management** - Create, read, update, delete DNS zones (master, slave, forward, stub)
- **Record Management** - Full CRUD for 20+ record types (A, AAAA, CNAME, MX, TXT, SRV, CAA, PTR, etc.)
- **DNSSEC Support** - Generate and manage DNSSEC keys (KSK, ZSK, CSK)
- **Data Sources** - Query existing zones and records
- **Import Support** - Import existing resources into Terraform state
- **Multi-Server** - Manage multiple BIND9 servers using provider aliases
- **Bulk Generation** - Create many records using Terraform's `for_each` and `range()`

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
  - [Single Server](#architecture-1-single-server)
  - [Multi-Primary Servers](#architecture-2-multi-primary-servers)
- [Installation](#installation)
  - [Option 1: Terraform Registry (Coming Soon)](#option-1-terraform-registry-coming-soon)
  - [Option 2: GitHub Releases ✅](#option-2-download-from-github-releases--currently-available)
  - [Option 3: Build from Source](#option-3-build-from-source)
- [Quick Start](#quick-start)
- [Single Server Setup](#single-server-setup)
- [Multi-Server Setup](#multi-server-setup)
- [Bulk Record Generation](#bulk-record-generation)
- [Resources](#resources)
- [Data Sources](#data-sources)
- [Supported Record Types](#supported-record-types)
- [Provider Arguments](#provider-arguments)
- [Import](#import)
- [Documentation](#documentation)

---

## Installation

### Option 1: Terraform Registry (Coming Soon)

> ⏳ **Not yet available.** This option will be available after the provider is published to the Terraform Registry.

Once published, this will be the easiest way - no manual download required:

```terraform
terraform {
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

provider "bind9" {
  endpoint = "http://your-bind9-server:8080"
  api_key  = var.bind9_api_key
}
```

Then simply run `terraform init` and Terraform will download the provider automatically.

---

### Option 2: Download from GitHub Releases ✅ (Currently Available)

Download pre-built binaries from GitHub releases.

#### Supported Platforms

| OS | Architecture | File |
|----|--------------|------|
| Linux | amd64 (x86_64) | `terraform-provider-bind9_1.0.0_linux_amd64.zip` |
| Linux | arm64 | `terraform-provider-bind9_1.0.0_linux_arm64.zip` |
| macOS | amd64 (Intel) | `terraform-provider-bind9_1.0.0_darwin_amd64.zip` |
| macOS | arm64 (Apple Silicon) | `terraform-provider-bind9_1.0.0_darwin_arm64.zip` |
| Windows | amd64 | `terraform-provider-bind9_1.0.0_windows_amd64.zip` |
| FreeBSD | amd64 | `terraform-provider-bind9_1.0.0_freebsd_amd64.zip` |

#### Step 1: Download

Download from: **https://github.com/harutyundermenjyan/terraform-provider-bind9/releases**

Or use command line:

**Linux (amd64):**
```bash
VERSION="1.0.0"
curl -LO "https://github.com/harutyundermenjyan/terraform-provider-bind9/releases/download/v${VERSION}/terraform-provider-bind9_${VERSION}_linux_amd64.zip"
unzip terraform-provider-bind9_${VERSION}_linux_amd64.zip
```

**macOS (Apple Silicon):**
```bash
VERSION="1.0.0"
curl -LO "https://github.com/harutyundermenjyan/terraform-provider-bind9/releases/download/v${VERSION}/terraform-provider-bind9_${VERSION}_darwin_arm64.zip"
unzip terraform-provider-bind9_${VERSION}_darwin_arm64.zip
```

**macOS (Intel):**
```bash
VERSION="1.0.0"
curl -LO "https://github.com/harutyundermenjyan/terraform-provider-bind9/releases/download/v${VERSION}/terraform-provider-bind9_${VERSION}_darwin_amd64.zip"
unzip terraform-provider-bind9_${VERSION}_darwin_amd64.zip
```

#### Step 2: Install

**Linux (amd64):**
```bash
mkdir -p ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/linux_amd64/
mv terraform-provider-bind9_v1.0.0 ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/linux_amd64/terraform-provider-bind9
chmod +x ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/linux_amd64/terraform-provider-bind9
```

**macOS (Apple Silicon):**
```bash
mkdir -p ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_arm64/
mv terraform-provider-bind9_v1.0.0 ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_arm64/terraform-provider-bind9
chmod +x ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_arm64/terraform-provider-bind9
```

**macOS (Intel):**
```bash
mkdir -p ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_amd64/
mv terraform-provider-bind9_v1.0.0 ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_amd64/terraform-provider-bind9
chmod +x ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/darwin_amd64/terraform-provider-bind9
```

#### Step 3: Configure Terraform

```terraform
terraform {
  required_providers {
    bind9 = {
      source  = "github.com/harutyundermenjyan/bind9"
      version = "1.0.0"
    }
  }
}

provider "bind9" {
  endpoint = "http://your-bind9-server:8080"
  api_key  = var.bind9_api_key
}
```

Then run:
```bash
terraform init
```

---

### Option 3: Build from Source

If you have [Go 1.21+](https://go.dev/dl/) installed:

```bash
git clone https://github.com/harutyundermenjyan/terraform-provider-bind9.git
cd terraform-provider-bind9
go build -o terraform-provider-bind9

# Install (auto-detects OS/arch)
mkdir -p ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/$(go env GOOS)_$(go env GOARCH)/
cp terraform-provider-bind9 ~/.terraform.d/plugins/github.com/harutyundermenjyan/bind9/1.0.0/$(go env GOOS)_$(go env GOARCH)/
```

---

## Quick Start

### Provider Configuration

```terraform
provider "bind9" {
  endpoint = "http://localhost:8080"
  api_key  = var.bind9_api_key
}
```

Or use environment variables:

```bash
export BIND9_ENDPOINT="http://localhost:8080"
export BIND9_API_KEY="your-api-key-here"
```

---

## Single Server Setup

Complete example for managing a single BIND9 server.

### File Structure

```
my-dns/
├── providers.tf      # Provider configuration
├── variables.tf      # Variables
├── zones.tf          # Zone definitions
├── records.tf        # DNS records
└── terraform.tfvars  # Your values (gitignored)
```

### providers.tf

```terraform
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
  endpoint = var.bind9_endpoint
  api_key  = var.bind9_api_key
}
```

### variables.tf

```terraform
variable "bind9_endpoint" {
  description = "BIND9 API endpoint URL"
  type        = string
  default     = "http://localhost:8080"
}

variable "bind9_api_key" {
  description = "BIND9 API key"
  type        = string
  sensitive   = true
}
```

### zones.tf

```terraform
resource "bind9_zone" "example" {
  name        = "example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 3600
  soa_retry   = 600
  soa_expire  = 604800
  soa_minimum = 86400
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]

  # Glue records for in-zone nameservers
  ns_addresses = {
    "ns1.example.com" = "10.0.0.1"
    "ns2.example.com" = "10.0.0.2"
  }

  allow_update           = ["key ddns-key"]
  delete_file_on_destroy = true
}

# Reverse DNS zone
resource "bind9_zone" "reverse" {
  name        = "0.0.10.in-addr.arpa"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]

  ns_addresses = {
    "ns1.example.com" = "10.0.0.1"
    "ns2.example.com" = "10.0.0.2"
  }
}
```

### records.tf

```terraform
# A Records
resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.0.100"]
}

resource "bind9_record" "mail" {
  zone    = bind9_zone.example.name
  name    = "mail"
  type    = "A"
  ttl     = 300
  records = ["10.0.0.50"]
}

# CNAME Record
resource "bind9_record" "api" {
  zone    = bind9_zone.example.name
  name    = "api"
  type    = "CNAME"
  ttl     = 300
  records = ["www.example.com."]
}

# MX Record
resource "bind9_record" "mx" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.example.com.", "20 mail2.example.com."]
}

# TXT Record (SPF)
resource "bind9_record" "spf" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 mx -all"]
}

# SRV Record
resource "bind9_record" "sip" {
  zone    = bind9_zone.example.name
  name    = "_sip._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["10 60 5060 sip.example.com."]
}

# CAA Record
resource "bind9_record" "caa" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "CAA"
  ttl     = 3600
  records = ["0 issue \"letsencrypt.org\""]
}

# PTR Record (Reverse DNS)
resource "bind9_record" "ptr_www" {
  zone    = bind9_zone.reverse.name
  name    = "100"
  type    = "PTR"
  ttl     = 300
  records = ["www.example.com."]
}
```

### terraform.tfvars

```terraform
bind9_endpoint = "http://localhost:8080"
bind9_api_key  = "your-api-key-here"
```

---

## Multi-Server Setup

For managing multiple independent BIND9 servers - **define records once, deploy to all or selected servers**.

### Architecture

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   Terraform/    │──────▶│   BIND9 API     │──────▶│     BIND9       │
│   OpenTofu      │       │   (dns1:8080)   │       │   Server 1      │
│                 │       └─────────────────┘       └─────────────────┘
│                 │
│                 │       ┌─────────────────┐       ┌─────────────────┐
│                 │──────▶│   BIND9 API     │──────▶│     BIND9       │
│                 │       │   (dns2:8080)   │       │   Server 2      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
```

### File Structure

```
bind9-orchestrator/
├── providers.tf          # Provider aliases for each server
├── variables.tf          # Server definitions
├── locals.tf             # Record definitions & expansion
├── zones.tf              # Zone resources per server
├── records.tf            # Records with server targeting
├── outputs.tf            # Outputs
└── terraform.tfvars      # Your server configuration (gitignored)
```

### providers.tf

```terraform
terraform {
  required_version = ">= 1.0"

  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# Provider for each DNS server
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

# Default provider (required by OpenTofu)
provider "bind9" {
  endpoint = var.servers["dns1"].endpoint
  api_key  = var.servers["dns1"].api_key
}
```

### variables.tf

```terraform
variable "servers" {
  description = "Map of BIND9 servers to manage"
  type = map(object({
    endpoint = string
    api_key  = string
    enabled  = bool
  }))
}
```

### locals.tf

```terraform
locals {
  # Filter to only enabled servers
  enabled_servers = {
    for name, server in var.servers : name => server
    if server.enabled
  }

  # ==========================================================================
  # Define records ONCE - deploy to ALL or SELECTED servers
  # ==========================================================================
  # servers = []               → deploy to ALL enabled servers
  # servers = ["dns1"]         → deploy to dns1 only
  # servers = ["dns1", "dns2"] → deploy to both dns1 and dns2
  # ==========================================================================

  example_com_records = {
    # A Records
    "www_A" = {
      name    = "www"
      type    = "A"
      ttl     = 300
      records = ["10.0.0.100"]
      servers = []              # ALL servers
    }
    "app_A" = {
      name    = "app"
      type    = "A"
      ttl     = 300
      records = ["10.0.0.101"]
      servers = []              # ALL servers
    }
    "db_A" = {
      name    = "db"
      type    = "A"
      ttl     = 300
      records = ["10.0.0.102"]
      servers = ["dns1"]        # dns1 only (internal)
    }
    "staging_A" = {
      name    = "staging"
      type    = "A"
      ttl     = 300
      records = ["10.0.0.200"]
      servers = ["dns2"]        # dns2 only (staging env)
    }

    # CNAME Records
    "api_CNAME" = {
      name    = "api"
      type    = "CNAME"
      ttl     = 300
      records = ["app.example.com."]
      servers = []
    }

    # MX Records
    "mx_MX" = {
      name    = "@"
      type    = "MX"
      ttl     = 3600
      records = ["10 mail.example.com."]
      servers = []
    }

    # TXT Records
    "spf_TXT" = {
      name    = "@"
      type    = "TXT"
      ttl     = 3600
      records = ["v=spf1 mx ~all"]
      servers = []
    }
  }

  # Expand records to target servers
  example_com_records_expanded = merge([
    for record_key, record in local.example_com_records : {
      for server_name, server in local.enabled_servers :
      "${record_key}_${server_name}" => merge(record, { server = server_name })
      if length(record.servers) == 0 || contains(record.servers, server_name)
    }
  ]...)
}
```

### zones.tf

```terraform
# Zone on dns1
resource "bind9_zone" "example_dns1" {
  count    = try(var.servers["dns1"].enabled, false) ? 1 : 0
  provider = bind9.dns1

  name        = "example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 3600
  soa_retry   = 600
  soa_expire  = 604800
  soa_minimum = 86400
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]

  ns_addresses = {
    "ns1.example.com" = "10.0.0.1"
    "ns2.example.com" = "10.0.0.2"
  }

  allow_update           = ["key ddns-key"]
  delete_file_on_destroy = true
}

# Zone on dns2
resource "bind9_zone" "example_dns2" {
  count    = try(var.servers["dns2"].enabled, false) ? 1 : 0
  provider = bind9.dns2

  name        = "example.com"
  type        = "master"
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 3600
  soa_retry   = 600
  soa_expire  = 604800
  soa_minimum = 86400
  default_ttl = 3600

  nameservers = ["ns1.example.com", "ns2.example.com"]

  ns_addresses = {
    "ns1.example.com" = "10.0.0.1"
    "ns2.example.com" = "10.0.0.2"
  }

  allow_update           = ["key ddns-key"]
  delete_file_on_destroy = true
}
```

### records.tf

```terraform
# Records for dns1
resource "bind9_record" "example_dns1" {
  for_each = {
    for k, v in local.example_com_records_expanded : k => v
    if v.server == "dns1"
  }
  provider = bind9.dns1

  zone    = bind9_zone.example_dns1[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}

# Records for dns2
resource "bind9_record" "example_dns2" {
  for_each = {
    for k, v in local.example_com_records_expanded : k => v
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

### outputs.tf

```terraform
output "zones" {
  description = "Created zones per server"
  value = {
    dns1 = try(bind9_zone.example_dns1[0].name, null)
    dns2 = try(bind9_zone.example_dns2[0].name, null)
  }
}

output "record_count" {
  description = "Number of records created per server"
  value = {
    dns1 = length(bind9_record.example_dns1)
    dns2 = length(bind9_record.example_dns2)
  }
}
```

### terraform.tfvars

```terraform
servers = {
  "dns1" = {
    endpoint = "http://localhost:8080"   # SSH tunnel or direct
    api_key  = "your-api-key-for-dns1"
    enabled  = true
  }
  "dns2" = {
    endpoint = "http://localhost:8081"
    api_key  = "your-api-key-for-dns2"
    enabled  = true
  }
}
```

### Using SSH Tunnels

If the BIND9 API is not directly accessible:

```bash
# Terminal 1: SSH tunnel to dns1
ssh -L 8080:localhost:8080 user@dns1.example.com

# Terminal 2: SSH tunnel to dns2
ssh -L 8081:localhost:8080 user@dns2.example.com

# Terminal 3: Run Terraform
tofu apply
```

---

## Bulk Record Generation

### $GENERATE Equivalent

BIND9's `$GENERATE` directive creates multiple similar records. Use Terraform's `range()` for the same result:

```terraform
locals {
  # Equivalent to BIND9: $GENERATE 1-50 host-$ A 10.0.1.$
  generated_hosts = {
    for i in range(1, 51) :
    "host-${i}_A" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.1.${i}"]
      servers = []
    }
  }

  # With step - equivalent to: $GENERATE 0-100/2 even-$ A 10.0.2.$
  generated_even = {
    for i in range(0, 101, 2) :  # start=0, end=101 (exclusive), step=2
    "even-${i}_A" => {
      name    = "even-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
      servers = []
    }
  }

  # Reverse PTR records - equivalent to: $GENERATE 1-254 $ PTR host-$.example.com.
  generated_ptr = {
    for i in range(1, 255) :
    "${i}_PTR" => {
      name    = "${i}"
      type    = "PTR"
      ttl     = 3600
      records = ["host-${i}.example.com."]
      servers = []
    }
  }
}

resource "bind9_record" "generated" {
  for_each = local.generated_hosts

  zone    = bind9_zone.example.name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

### Dynamic Records from Lists

```terraform
variable "web_servers" {
  type = map(string)
  default = {
    "web1" = "10.0.0.101"
    "web2" = "10.0.0.102"
    "web3" = "10.0.0.103"
    "web4" = "10.0.0.104"
  }
}

resource "bind9_record" "web_servers" {
  for_each = var.web_servers

  zone    = bind9_zone.example.name
  name    = each.key
  type    = "A"
  ttl     = 300
  records = [each.value]
}
```

---

## DNSSEC

### Enable DNSSEC on a Zone

```terraform
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.example.name
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
}

resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.example.name
  key_type  = "ZSK"
  algorithm = 13
  sign_zone = true  # Sign the zone after key creation
}

# DS records to submit to your domain registrar
output "ds_records" {
  value = bind9_dnssec_key.ksk.ds_records
}
```

---

## Resources

| Resource | Description |
|----------|-------------|
| [`bind9_zone`](docs/resources/zone.md) | Manages DNS zones (master, slave, forward, stub) |
| [`bind9_record`](docs/resources/record.md) | Manages DNS records (A, AAAA, CNAME, MX, TXT, etc.) |
| [`bind9_dnssec_key`](docs/resources/dnssec_key.md) | Manages DNSSEC keys (KSK, ZSK, CSK) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [`bind9_zone`](docs/data-sources/zone.md) | Query a specific zone |
| [`bind9_zones`](docs/data-sources/zones.md) | List all zones |
| [`bind9_record`](docs/data-sources/record.md) | Query a specific record |
| [`bind9_records`](docs/data-sources/records.md) | List records in a zone |

### Query Examples

```terraform
# Get zone info
data "bind9_zone" "info" {
  name = "example.com"
}

# List all zones
data "bind9_zones" "all" {}

# Get specific record
data "bind9_record" "www" {
  zone = "example.com"
  name = "www"
  type = "A"
}

# List all records in zone
data "bind9_records" "all" {
  zone = "example.com"
}

# Filter records by type
data "bind9_records" "mx_records" {
  zone = "example.com"
  type = "MX"
}
```

## Supported Record Types

| Type | Description | Example |
|------|-------------|---------|
| `A` | IPv4 address | `["10.0.0.100"]` |
| `AAAA` | IPv6 address | `["2001:db8::1"]` |
| `CNAME` | Canonical name (alias) | `["www.example.com."]` |
| `MX` | Mail exchange | `["10 mail.example.com."]` |
| `TXT` | Text record | `["v=spf1 mx -all"]` |
| `NS` | Nameserver | `["ns1.example.com."]` |
| `PTR` | Pointer (reverse DNS) | `["www.example.com."]` |
| `SRV` | Service locator | `["10 60 5060 sip.example.com."]` |
| `CAA` | Certificate Authority | `["0 issue \"letsencrypt.org\""]` |
| `NAPTR` | Name Authority Pointer | `["100 10 \"\" \"\" \"!^.*$!sip:info@example.com!\" ."]` |
| `SSHFP` | SSH fingerprint | `["2 1 123456..."]` |
| `TLSA` | DANE/TLS Association | `["3 1 1 abc123..."]` |
| `DNAME` | Delegation name | `["other.example.com."]` |
| `LOC` | Geographic location | `["52 22 23.000 N 4 53 32.000 E 0.00m"]` |
| `HTTPS` | HTTPS binding | `["1 . alpn=\"h2,h3\""]` |
| `SVCB` | Service binding | `["1 . alpn=\"h2\""]` |
| `HINFO` | Host information | `["Intel Linux"]` |
| `RP` | Responsible person | `["admin.example.com. ."]` |
| `URI` | Uniform Resource Identifier | `["10 1 \"https://example.com/\""]` |
| `DNSKEY` | DNSSEC public key | (managed by DNSSEC resources) |
| `DS` | Delegation Signer | (managed by DNSSEC resources) |

## Provider Arguments

| Argument | Description | Default | Env Var |
|----------|-------------|---------|---------|
| `endpoint` | BIND9 API URL | Required | `BIND9_ENDPOINT` |
| `api_key` | API key for authentication | - | `BIND9_API_KEY` |
| `username` | Username for JWT auth | - | `BIND9_USERNAME` |
| `password` | Password for JWT auth | - | `BIND9_PASSWORD` |
| `insecure` | Skip TLS certificate verification | `false` | - |
| `timeout` | Request timeout in seconds | `30` | - |

## Import

Import existing resources into Terraform state:

```bash
# Import a zone
terraform import bind9_zone.example example.com

# Import a record (format: zone/name/type)
terraform import bind9_record.www example.com/www/A
terraform import bind9_record.mx "example.com/@/MX"

# DNSSEC keys cannot be imported (security reasons)
```

---

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/harutyundermenjyan/bind9/latest/docs).

- [Getting Started](docs/guides/getting-started.md)
- [Provider Configuration](docs/index.md)
- [bind9_zone Resource](docs/resources/zone.md)
- [bind9_record Resource](docs/resources/record.md)
- [bind9_dnssec_key Resource](docs/resources/dnssec_key.md)
- [bind9_zone Data Source](docs/data-sources/zone.md)
- [bind9_zones Data Source](docs/data-sources/zones.md)
- [bind9_record Data Source](docs/data-sources/record.md)
- [bind9_records Data Source](docs/data-sources/records.md)

---

## Related Projects

| Project | Description |
|---------|-------------|
| **[bind9-api](https://github.com/harutyundermenjyan/bind9-api)** | BIND9 REST API server (**required**) |

---

## Requirements

| Requirement | Description |
|-------------|-------------|
| Terraform >= 1.0 or OpenTofu >= 1.0 | Infrastructure as Code tool |
| **[BIND9 REST API](https://github.com/harutyundermenjyan/bind9-api)** | **Required** - Must be installed on each BIND9 server |
| BIND9 9.x | DNS server with rndc and nsupdate |

---

## Building from Source

```bash
git clone https://github.com/harutyundermenjyan/terraform-provider-bind9.git
cd terraform-provider-bind9
go build -o terraform-provider-bind9
```

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o terraform-provider-bind9

# Install locally for testing
mkdir -p ~/.terraform.d/plugins/local/bind9/bind9/1.0.0/$(go env GOOS)_$(go env GOARCH)
cp terraform-provider-bind9 ~/.terraform.d/plugins/local/bind9/bind9/1.0.0/$(go env GOOS)_$(go env GOARCH)/
```

---

## Author

**Harutyun Dermenjyan**

- GitHub: [@harutyundermenjyan](https://github.com/harutyundermenjyan)

## License

MIT License - see [LICENSE](LICENSE) for details.

Copyright © 2024 Harutyun Dermenjyan
