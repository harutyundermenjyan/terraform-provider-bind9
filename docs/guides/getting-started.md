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

1. **BIND9 Server** with the **BIND9 REST API** installed
2. **API Key** for authentication
3. **Terraform/OpenTofu** version 1.0 or later

> **Important:** This provider requires the [BIND9 REST API](https://github.com/harutyundermenjyan/bind9-api) to be installed and configured on your BIND9 server. See the [BIND9 Server Setup](#bind9-server-setup) section below.

---

## BIND9 Server Setup

### Required Components

| Component | Repository | Description |
|-----------|------------|-------------|
| **BIND9** | System package | DNS server |
| **BIND9 REST API** | [github.com/harutyundermenjyan/bind9-api](https://github.com/harutyundermenjyan/bind9-api) | REST API that this provider communicates with |

### Architecture

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│   Terraform/    │──────▶│   BIND9 REST    │──────▶│     BIND9       │
│   OpenTofu      │ HTTP  │      API        │ rndc  │     Server      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
     (this provider)        (:8080)               (DNS on :53)
```

### Quick Setup (Ubuntu/Debian)

#### Step 1: Install BIND9

```bash
apt update
apt install -y bind9 bind9utils
```

#### Step 2: Create Required Directories

```bash
mkdir -p /etc/bind/keys
mkdir -p /var/lib/bind
mkdir -p /var/log/bind

chown -R bind:bind /etc/bind/keys /var/lib/bind /var/log/bind
```

#### Step 3: Generate RNDC and TSIG Keys

```bash
# RNDC key for server control
rndc-confgen -a -k rndc-key -c /etc/bind/rndc.key
chown bind:bind /etc/bind/rndc.key
chmod 640 /etc/bind/rndc.key

# TSIG key for dynamic DNS updates
tsig-keygen -a hmac-sha256 ddns-key > /etc/bind/keys/ddns-key.key
chown bind:bind /etc/bind/keys/ddns-key.key
chmod 640 /etc/bind/keys/ddns-key.key
```

#### Step 4: Configure BIND9

Edit `/etc/bind/named.conf`:

```bind
// Include keys first
include "/etc/bind/rndc.key";
include "/etc/bind/keys/ddns-key.key";

// Include API-managed ACLs (for bind9_acl resource)
include "/etc/bind/named.conf.acls";

// Then includes
include "/etc/bind/named.conf.options";
include "/etc/bind/named.conf.local";
include "/etc/bind/named.conf.default-zones";
```

Edit `/etc/bind/named.conf.options`:

```bind
options {
    directory "/var/cache/bind";

    allow-query { any; };
    dnssec-validation auto;
    listen-on { any; };
    listen-on-v6 { any; };

    // CRITICAL: Required for API zone management
    allow-new-zones yes;
};

// Statistics channel for API
statistics-channels {
    inet 127.0.0.1 port 8053 allow { 127.0.0.1; };
};

// RNDC control
controls {
    inet 127.0.0.1 port 953 allow { 127.0.0.1; } keys { "rndc-key"; };
};
```

Create the ACL file (for `bind9_acl` resource):

```bash
touch /etc/bind/named.conf.acls
chown bind:bind /etc/bind/named.conf.acls
chmod 664 /etc/bind/named.conf.acls
```

#### Step 5: Install BIND9 REST API

```bash
cd /opt
git clone https://github.com/harutyundermenjyan/bind9-api.git
cd bind9-api

# Create virtual environment
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# Generate API key
python3 -c "import secrets; print('API_KEY:', secrets.token_urlsafe(32))"
# SAVE THIS KEY - you'll need it for Terraform!
```

#### Step 6: Configure API

Create `/opt/bind9-api/.env`:

```bash
# API Server
BIND9_API_HOST=0.0.0.0
BIND9_API_PORT=8080

# Authentication
BIND9_API_AUTH_ENABLED=true
BIND9_API_AUTH_STATIC_API_KEY=<your-generated-api-key>
BIND9_API_AUTH_STATIC_API_KEY_SCOPES=read,write,admin,dnssec,stats

# Database disabled (using static API key)
BIND9_API_DATABASE_ENABLED=false

# BIND9 Paths
BIND9_API_BIND9_CONFIG_PATH=/etc/bind/named.conf
BIND9_API_BIND9_ZONES_PATH=/var/lib/bind
BIND9_API_BIND9_KEYS_PATH=/etc/bind/keys
BIND9_API_BIND9_RNDC_KEY=/etc/bind/rndc.key
BIND9_API_BIND9_ACLS_PATH=/etc/bind/named.conf.acls
BIND9_API_BIND9_NAMED_CHECKZONE=/usr/bin/named-checkzone
BIND9_API_BIND9_NAMED_CHECKCONF=/usr/bin/named-checkconf

# TSIG Key (from /etc/bind/keys/ddns-key.key)
BIND9_API_TSIG_KEY_FILE=/etc/bind/keys/ddns-key.key
BIND9_API_TSIG_KEY_NAME=ddns-key
BIND9_API_TSIG_KEY_SECRET=<secret-from-ddns-key.key>
BIND9_API_TSIG_KEY_ALGORITHM=hmac-sha256

# Misc
BIND9_API_LOG_LEVEL=INFO
```

#### Step 7: Create Systemd Service

Create `/etc/systemd/system/bind9-api.service`:

```ini
[Unit]
Description=BIND9 REST API
After=network.target named.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/bind9-api
EnvironmentFile=/opt/bind9-api/.env
ExecStart=/opt/bind9-api/venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8080
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable bind9-api
systemctl start bind9-api
```

#### Step 8: Verify Setup

```bash
# Check BIND9
rndc status

# Check API
curl http://localhost:8080/health

# Test authentication
curl -H "X-API-Key: YOUR_API_KEY" http://localhost:8080/api/v1/zones
```

### File Permissions Reference

| File | Owner | Permissions | Purpose |
|------|-------|-------------|---------|
| `/etc/bind/named.conf` | `root:bind` | `644` | Main BIND9 config |
| `/etc/bind/named.conf.acls` | `bind:bind` | `664` | API-managed ACLs |
| `/etc/bind/rndc.key` | `bind:bind` | `640` | RNDC authentication |
| `/etc/bind/keys/ddns-key.key` | `bind:bind` | `640` | TSIG key for updates |
| `/var/lib/bind/` | `bind:bind` | `755` | Zone files directory |
| `/opt/bind9-api/.env` | `root:root` | `600` | API configuration |

> **Full Setup Guide:** See the [BIND9 REST API Setup Guide](https://github.com/harutyundermenjyan/bind9-api/blob/main/SETUP.md) for complete instructions including HTTPS, AppArmor, and troubleshooting.

---

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

### MX Record Creation Fails with "REFUSED"

```
Error: Could not create record @ MX: API error 500: update failed: REFUSED
```

**Cause**: BIND9 validates that MX record targets have A/AAAA records before allowing the MX record to be created.

**Solution**: Always create the mail server A record BEFORE the MX record using `depends_on`:

```hcl
# Mail server A record - MUST be created first
resource "bind9_record" "mail" {
  zone    = bind9_zone.main.name
  name    = "mail"
  type    = "A"
  ttl     = 300
  records = ["172.25.44.101"]
}

# MX record - depends on mail A record
resource "bind9_record" "mx" {
  zone    = bind9_zone.main.name
  name    = "@"
  type    = "MX"
  ttl     = 3600
  records = ["10 mail.example.com."]
  
  depends_on = [bind9_record.mail]  # Required!
}
```

