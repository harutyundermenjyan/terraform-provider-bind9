# Terraform Provider for BIND9

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Terraform Registry](https://img.shields.io/badge/Terraform-Registry-purple.svg)](https://registry.terraform.io/providers/harutyundermenjyan/bind9)

A Terraform/OpenTofu provider for managing DNS zones and records on BIND9 servers via the BIND9 REST API.

## Author

**Harutyun Dermenjyan**

- GitHub: [@harutyundermenjyan](https://github.com/harutyundermenjyan)
- Repository: [terraform-provider-bind9](https://github.com/harutyundermenjyan/terraform-provider-bind9)

## Related Projects

| Project | Description |
|---------|-------------|
| [bind9-api](https://gitlab.com/Dermenjyan/bind9-api) | REST API for BIND9 DNS Server |
| [bind9-orchestrator](https://gitlab.com/Dermenjyan/bind9-orchestrator) | Terraform configurations for multi-server DNS management |

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}
```

### From Source

```bash
git clone https://github.com/harutyundermenjyan/terraform-provider-bind9.git
cd terraform-provider-bind9
go build -o terraform-provider-bind9
```

## Documentation

Full documentation is available in the [docs/](./docs/) directory and on the [Terraform Registry](https://registry.terraform.io/providers/harutyundermenjyan/bind9/latest/docs).

- **[Provider Configuration](./docs/index.md)** - Authentication and setup
- **Resources:**
  - [bind9_zone](./docs/resources/zone.md) - Manage DNS zones
  - [bind9_record](./docs/resources/record.md) - Manage DNS records
  - [bind9_dnssec_key](./docs/resources/dnssec_key.md) - Manage DNSSEC keys
- **Data Sources:**
  - [bind9_zone](./docs/data-sources/zone.md) - Query zone information
  - [bind9_zones](./docs/data-sources/zones.md) - List all zones
  - [bind9_record](./docs/data-sources/record.md) - Query specific records
  - [bind9_records](./docs/data-sources/records.md) - List records in a zone

## Features

- **Zone Management**: Create, read, update, delete DNS zones
- **Record Management**: Full CRUD for all DNS record types (A, AAAA, CNAME, MX, TXT, SRV, CAA, NS, PTR, etc.)
- **DNSSEC**: Key generation, signing, DS record output
- **Data Sources**: Query existing zones and records
- **Import Support**: Import existing zones and records into Terraform state

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 or [OpenTofu](https://opentofu.org/) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building)
- BIND9 REST API server running ([bind9-api](https://gitlab.com/Dermenjyan/bind9-api))

## Quick Start

### Provider Configuration

```hcl
provider "bind9" {
  endpoint = "http://localhost:8080"
  api_key  = var.bind9_api_key
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `BIND9_ENDPOINT` | API endpoint URL |
| `BIND9_API_KEY` | API key for authentication |
| `BIND9_USERNAME` | Username for JWT auth |
| `BIND9_PASSWORD` | Password for JWT auth |

### Create a Zone

```hcl
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
  
  allow_update = ["key ddns-key"]
}
```

### Create Records

```hcl
# A Record
resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.2.100"]
}

# MX Record
resource "bind9_record" "mx" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.example.com."]
}

# TXT Record (SPF)
resource "bind9_record" "spf" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 mx ~all"]
}
```

## Supported Record Types

A, AAAA, CNAME, MX, TXT, NS, PTR, SOA, SRV, CAA, NAPTR, HTTPS, SVCB, TLSA, SSHFP, DNSKEY, DS, LOC, HINFO, RP, DNAME, URI

## Import

```bash
# Import Zone
terraform import bind9_zone.example example.com

# Import Record (format: zone/name/type)
terraform import bind9_record.www example.com/www/A
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

MIT License - Copyright (c) 2024 Harutyun Dermenjyan

See [LICENSE](LICENSE) for details.
