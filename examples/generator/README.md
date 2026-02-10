# BIND9 Terraform Configuration Generator

Generates ready-to-use Terraform/OpenTofu configurations for managing BIND9 DNS servers.

Define your servers, zones, records, and ACLs in a simple `config.yaml`, then run the generator to produce a complete Terraform project.

## Quick Start

```bash
# Install dependency
pip install pyyaml

# Edit config.yaml with your settings
vim config.yaml

# Generate Terraform configurations
python generate.py

# Deploy
cd generated
tofu init
tofu plan
tofu apply
```

## Configuration

Edit `config.yaml` to define:

- **Servers** - Any number of BIND9 servers (dynamic, not hardcoded)
- **ACLs** - Access control lists with optional server targeting
- **Zones** - DNS zones with SOA defaults
- **Records** - DNS records (A, AAAA, CNAME, MX, TXT, SRV, CAA, and more)

### Server Targeting

Every resource supports `servers: []` for targeting:

```yaml
# Deploy to ALL servers
servers: []

# Deploy only to specific servers
servers: ["dns1"]
```

## Generated Files

```
generated/
├── versions.tf       # Provider requirements
├── variables.tf      # Server variable definitions
├── terraform.tfvars  # Server credentials (gitignored)
├── acls.tf           # ACL definitions
├── zones.tf          # Zone definitions with SOA defaults
├── records.tf        # Record definitions
├── servers.tf        # Provider aliases and resources per server
└── .gitignore        # Ignores sensitive files
```

## Usage

```bash
# Default config
python generate.py

# Custom config file
python generate.py -c production.yaml

# Custom output directory
python generate.py -o ./production

# Dry run (preview only)
python generate.py --dry-run
```

## Adding More Servers

Just add servers to `config.yaml` and re-generate:

```yaml
servers:
  - name: dns1
    endpoint: "http://dns1:8080"
    api_key: "key-1"
    enabled: true
  - name: dns2
    endpoint: "http://dns2:8080"
    api_key: "key-2"
    enabled: true
  - name: dns3            # New server
    endpoint: "http://dns3:8080"
    api_key: "key-3"
    enabled: true
```

The generator creates provider aliases and resource blocks for each server automatically.
