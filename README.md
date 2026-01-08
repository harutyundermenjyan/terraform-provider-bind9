# Terraform Provider for BIND9

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-purple.svg)](https://registry.terraform.io/providers/harutyundermenjyan/bind9/latest)

A Terraform/OpenTofu provider for managing DNS zones and records on BIND9 servers via REST API.

## Features

- **Zone Management** - Create, read, update, delete DNS zones (master, slave, forward, stub)
- **Record Management** - Full CRUD for 20+ record types (A, AAAA, CNAME, MX, TXT, SRV, CAA, PTR, etc.)
- **DNSSEC Support** - Generate and manage DNSSEC keys (KSK, ZSK, CSK)
- **Data Sources** - Query existing zones and records
- **Import Support** - Import existing resources into Terraform state
- **Multi-Server** - Manage multiple BIND9 servers using provider aliases

## Quick Start

### Installation

```terraform
terraform {
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}
```

### Provider Configuration

```terraform
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  api_key  = var.bind9_api_key
}
```

Or use environment variables:

```bash
export BIND9_ENDPOINT="https://dns.example.com:8080"
export BIND9_API_KEY="your-api-key-here"
```

### Create a Zone

```terraform
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
}
```

### Create Records

```terraform
# A Record
resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
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
```

### Enable DNSSEC

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
  sign_zone = true
}

# DS records to submit to registrar
output "ds_records" {
  value = bind9_dnssec_key.ksk.ds_records
}
```

### Query Existing Resources

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
```

## Resources

| Resource | Description |
|----------|-------------|
| `bind9_zone` | Manages DNS zones (master, slave, forward, stub) |
| `bind9_record` | Manages DNS records (A, AAAA, CNAME, MX, TXT, etc.) |
| `bind9_dnssec_key` | Manages DNSSEC keys (KSK, ZSK, CSK) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `bind9_zone` | Query a specific zone |
| `bind9_zones` | List all zones |
| `bind9_record` | Query a specific record |
| `bind9_records` | List records in a zone |

## Supported Record Types

| Type | Description |
|------|-------------|
| `A` | IPv4 address |
| `AAAA` | IPv6 address |
| `CNAME` | Canonical name (alias) |
| `MX` | Mail exchange |
| `TXT` | Text record |
| `NS` | Nameserver |
| `PTR` | Pointer (reverse DNS) |
| `SRV` | Service locator |
| `CAA` | Certificate Authority Authorization |
| `NAPTR` | Name Authority Pointer |
| `SSHFP` | SSH fingerprint |
| `TLSA` | DANE/TLS Association |
| `DNAME` | Delegation name |
| `LOC` | Geographic location |
| `HTTPS` | HTTPS binding |
| `SVCB` | Service binding |
| `HINFO` | Host information |
| `RP` | Responsible person |
| `URI` | Uniform Resource Identifier |
| `DNSKEY` | DNSSEC public key |
| `DS` | Delegation Signer |

## Provider Arguments

| Argument | Description | Default |
|----------|-------------|---------|
| `endpoint` | BIND9 API URL (or `BIND9_ENDPOINT` env var) | Required |
| `api_key` | API key (or `BIND9_API_KEY` env var) | - |
| `username` | Username for JWT auth (or `BIND9_USERNAME` env var) | - |
| `password` | Password for JWT auth (or `BIND9_PASSWORD` env var) | - |
| `insecure` | Skip TLS verification | `false` |
| `timeout` | Request timeout in seconds | `30` |

## Import

Import existing resources:

```bash
# Import a zone
terraform import bind9_zone.example example.com

# Import a record (format: zone/name/type)
terraform import bind9_record.www example.com/www/A
terraform import bind9_record.mx "example.com/@/MX"
```

## Multi-Server Configuration

```terraform
provider "bind9" {
  alias    = "dns1"
  endpoint = "https://dns1.example.com:8080"
  api_key  = var.dns1_api_key
}

provider "bind9" {
  alias    = "dns2"
  endpoint = "https://dns2.example.com:8080"
  api_key  = var.dns2_api_key
}

resource "bind9_zone" "primary" {
  provider = bind9.dns1
  name     = "example.com"
  type     = "master"
  # ...
}

resource "bind9_zone" "secondary" {
  provider = bind9.dns2
  name     = "example.com"
  type     = "slave"
}
```

## Documentation

Full documentation is available on the [Terraform Registry](https://registry.terraform.io/providers/harutyundermenjyan/bind9/latest/docs).

- [Provider Configuration](docs/index.md)
- [bind9_zone Resource](docs/resources/zone.md)
- [bind9_record Resource](docs/resources/record.md)
- [bind9_dnssec_key Resource](docs/resources/dnssec_key.md)
- [bind9_zone Data Source](docs/data-sources/zone.md)
- [bind9_zones Data Source](docs/data-sources/zones.md)
- [bind9_record Data Source](docs/data-sources/record.md)
- [bind9_records Data Source](docs/data-sources/records.md)

## Related Projects

| Project | Description |
|---------|-------------|
| [bind9-api](https://gitlab.com/Dermenjyan/bind9-api) | BIND9 REST API server |
| [bind9-orchestrator](https://gitlab.com/Dermenjyan/bind9-orchestrator) | Example configurations |

## Requirements

- Terraform >= 1.0 or OpenTofu >= 1.0
- BIND9 REST API server

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

# Install locally
make install
```

## Author

**Harutyun Dermenjyan**

- GitHub: [@harutyundermenjyan](https://github.com/harutyundermenjyan)
- GitLab: [@Dermenjyan](https://gitlab.com/Dermenjyan)

## License

MIT License - see [LICENSE](LICENSE) for details.

Copyright © 2024 Harutyun Dermenjyan
