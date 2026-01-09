---
page_title: "BIND9 Provider"
description: |-
  The BIND9 provider enables Terraform/OpenTofu to manage DNS zones and records on BIND9 servers through a REST API.
---

# BIND9 Provider

The BIND9 provider allows you to manage DNS zones and records on a BIND9 server through its REST API. It supports zone management, DNS record operations, and DNSSEC key management.

## Features

- **Zone Management** - Create, update, and delete DNS zones (master, slave, forward, stub)
- **Record Management** - Full CRUD support for all common DNS record types
- **DNSSEC Support** - Generate and manage DNSSEC keys (KSK, ZSK, CSK)
- **Data Sources** - Query existing zones and records
- **Import Support** - Import existing resources into Terraform state
- **Multi-Server** - Manage multiple BIND9 servers using provider aliases

## Supported Record Types

| Record Type | Description | Example Use Case |
|-------------|-------------|------------------|
| `A` | IPv4 address | Web server address |
| `AAAA` | IPv6 address | IPv6-enabled services |
| `CNAME` | Canonical name (alias) | www → web-server |
| `MX` | Mail exchange | Email routing |
| `TXT` | Text record | SPF, DKIM, DMARC |
| `NS` | Nameserver | Zone delegation |
| `PTR` | Pointer (reverse DNS) | IP to hostname mapping |
| `SRV` | Service locator | LDAP, SIP, Kerberos |
| `CAA` | Certificate Authority Authorization | SSL certificate policy |
| `NAPTR` | Name Authority Pointer | ENUM, SIP |
| `SSHFP` | SSH fingerprint | SSH host verification |
| `TLSA` | DANE/TLS Association | Certificate pinning |
| `DNAME` | Delegation name | Subtree aliasing |
| `LOC` | Geographic location | Physical location |
| `HTTPS` | HTTPS binding | Service binding |
| `SVCB` | Service binding | General service binding |
| `HINFO` | Host information | OS and hardware info |
| `RP` | Responsible person | Contact information |
| `URI` | Uniform Resource Identifier | Service endpoints |

## Prerequisites

~> **Important:** This provider requires the **BIND9 REST API** to be installed on your BIND9 server(s).

### Required: BIND9 REST API

This provider communicates with BIND9 through a REST API - it does NOT connect directly to BIND9.

**You must install the BIND9 REST API first:**

| Component | Repository | Description |
|-----------|------------|-------------|
| **BIND9 REST API** | [github.com/harutyundermenjyan/bind9-api](https://github.com/harutyundermenjyan/bind9-api) | REST API server that runs on each BIND9 server |

### Architecture

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   Terraform/    │──────▶│   BIND9 REST    │──────▶│     BIND9       │
│   OpenTofu      │ HTTPS │      API        │ rndc  │     Server      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
     (this provider)      (required component)       (DNS server)
```

### Checklist

1. **BIND9 REST API** - Install and configure on each BIND9 server ([Installation Guide](https://github.com/harutyundermenjyan/bind9-api#quick-start))
2. **Authentication** - Generate an API key during API setup
3. **Network Access** - Ensure Terraform can reach the API endpoint (default port: 8080)

## Example Usage

### Basic Configuration

```terraform
terraform {
  required_providers {
    bind9 = {
      source  = "harutyundermenjyan/bind9"
      version = "~> 1.0"
    }
  }
}

# Configure the provider
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  api_key  = var.bind9_api_key
}

# Create a zone
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

# Create an A record
resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["10.0.1.100"]
}

# Create an MX record
resource "bind9_record" "mx" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.example.com."]
}
```

### Using Environment Variables

```terraform
provider "bind9" {
  # All values can be set via environment variables:
  # BIND9_ENDPOINT - API endpoint URL
  # BIND9_API_KEY  - API key for authentication
}
```

```bash
export BIND9_ENDPOINT="https://dns.example.com:8080"
export BIND9_API_KEY="your-api-key-here"
terraform apply
```

### Multiple BIND9 Servers

```terraform
# Primary DNS server
provider "bind9" {
  alias    = "primary"
  endpoint = "https://dns1.example.com:8080"
  api_key  = var.dns1_api_key
}

# Secondary DNS server
provider "bind9" {
  alias    = "secondary"
  endpoint = "https://dns2.example.com:8080"
  api_key  = var.dns2_api_key
}

# Zone on primary
resource "bind9_zone" "primary" {
  provider = bind9.primary
  
  name = "example.com"
  type = "master"
  # ...
}

# Zone on secondary
resource "bind9_zone" "secondary" {
  provider = bind9.secondary
  
  name = "example.com"
  type = "slave"
  # ...
}
```

### Complete Zone with DNSSEC

```terraform
resource "bind9_zone" "secure" {
  name = "secure.example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  
  nameservers = ["ns1.example.com", "ns2.example.com"]
  
  ns_addresses = {
    "ns1.example.com" = "10.0.1.10"
    "ns2.example.com" = "10.0.1.11"
  }
  
  allow_transfer = ["10.0.1.11"]
  notify         = true
}

# Enable DNSSEC
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.secure.name
  key_type  = "KSK"
  algorithm = 13
}

resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.secure.name
  key_type  = "ZSK"
  algorithm = 13
  sign_zone = true
}

# Output DS records for registrar
output "ds_records" {
  value = bind9_dnssec_key.ksk.ds_records
}
```

## Authentication

The provider supports two authentication methods:

### API Key (Recommended)

```terraform
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  api_key  = var.bind9_api_key  # or use BIND9_API_KEY env var
}
```

### Username/Password (JWT)

```terraform
provider "bind9" {
  endpoint = "https://dns.example.com:8080"
  username = var.bind9_username  # or use BIND9_USERNAME env var
  password = var.bind9_password  # or use BIND9_PASSWORD env var
}
```

## Schema

### Required

No required arguments if using environment variables.

### Optional

- `endpoint` (String) BIND9 REST API endpoint URL (e.g., `https://dns.example.com:8080`). Can also be set via `BIND9_ENDPOINT` environment variable.
- `api_key` (String, Sensitive) API key for authentication. Can also be set via `BIND9_API_KEY` environment variable.
- `username` (String) Username for JWT authentication. Can also be set via `BIND9_USERNAME` environment variable.
- `password` (String, Sensitive) Password for JWT authentication. Can also be set via `BIND9_PASSWORD` environment variable.
- `insecure` (Boolean) Skip TLS certificate verification. Default: `false`. Use only for testing.
- `timeout` (Number) API request timeout in seconds. Default: `30`.

## Resources

| Resource | Description |
|----------|-------------|
| [bind9_zone](resources/zone.md) | Manages a DNS zone on BIND9 server |
| [bind9_record](resources/record.md) | Manages DNS records on BIND9 server |
| [bind9_dnssec_key](resources/dnssec_key.md) | Manages DNSSEC keys for zones |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [bind9_zone](data-sources/zone.md) | Retrieves information about a specific zone |
| [bind9_zones](data-sources/zones.md) | Lists all zones with optional filtering |
| [bind9_record](data-sources/record.md) | Retrieves a specific record by name and type |
| [bind9_records](data-sources/records.md) | Lists all records in a zone with optional filtering |

## Import

All resources support importing existing infrastructure:

```bash
# Import a zone
terraform import bind9_zone.example example.com

# Import a record
terraform import bind9_record.www example.com/www/A

# Import uses format: zone/name/type for records
```

## Related Projects

| Project | Description |
|---------|-------------|
| [bind9-api](https://gitlab.com/Dermenjyan/bind9-api) | BIND9 REST API server that this provider communicates with |
| [bind9-orchestrator](https://gitlab.com/Dermenjyan/bind9-orchestrator) | Example Terraform/OpenTofu configurations for managing BIND9 infrastructure |

## Author

**Harutyun Dermenjyan**

- GitLab: [@Dermenjyan](https://gitlab.com/Dermenjyan)
- GitHub: [@harutyundermenjyan](https://github.com/harutyundermenjyan)

## License

This provider is released under the Apache License 2.0.
