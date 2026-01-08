---
page_title: "bind9_dnssec_key Resource - BIND9 Provider"
subcategory: "DNSSEC"
description: |-
  Manages DNSSEC keys for a zone on BIND9 server.
---

# bind9_dnssec_key (Resource)

Manages DNSSEC keys for DNS zones on a BIND9 server. Supports Key Signing Keys (KSK), Zone Signing Keys (ZSK), and Combined Signing Keys (CSK).

## Example Usage

### Key Signing Key (KSK)

```hcl
resource "bind9_dnssec_key" "ksk" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}
```

### Zone Signing Key (ZSK)

```hcl
resource "bind9_dnssec_key" "zsk" {
  zone      = "example.com"
  key_type  = "ZSK"
  algorithm = 13  # ECDSAP256SHA256
}
```

### Combined Signing Key (CSK)

```hcl
resource "bind9_dnssec_key" "csk" {
  zone      = "example.com"
  key_type  = "CSK"
  algorithm = 13
  sign_zone = true
}
```

### Complete DNSSEC Setup

```hcl
# Create the zone first
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
  
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
}

# Generate KSK (Key Signing Key)
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.example.name
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256 (recommended)
  ttl       = 3600
  sign_zone = true
}

# Generate ZSK (Zone Signing Key)
resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.example.name
  key_type  = "ZSK"
  algorithm = 13
  ttl       = 3600
  
  depends_on = [bind9_dnssec_key.ksk]
}

# Output DS records for your registrar
output "ds_records" {
  description = "DS records to add at your registrar"
  value       = bind9_dnssec_key.ksk.ds_records
}
```

### RSA Key with Custom Size

```hcl
resource "bind9_dnssec_key" "ksk_rsa" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 8     # RSASHA256
  bits      = 2048  # Key size for RSA
  sign_zone = true
}
```

## Argument Reference

### Required

- `zone` (String) - The zone name for this DNSSEC key. Changing this forces a new resource.
- `key_type` (String) - The type of DNSSEC key. Valid values:
  - `KSK` - Key Signing Key (signs DNSKEY records)
  - `ZSK` - Zone Signing Key (signs all other records)
  - `CSK` - Combined Signing Key (acts as both KSK and ZSK)

### Optional

- `algorithm` (Number) - DNSSEC algorithm number. Default: `13` (ECDSAP256SHA256). See [Algorithm Reference](#algorithm-reference).
- `bits` (Number) - Key size in bits. Only applicable for RSA algorithms (8, 10). ECDSA and EdDSA have fixed sizes.
- `ttl` (Number) - TTL for the DNSKEY record. Default: `3600`
- `sign_zone` (Boolean) - Sign the zone after key creation. Default: `false`

### Read-Only

- `id` (String) - Key identifier in format `zone/key_tag`.
- `key_tag` (Number) - The DNSSEC key tag (unique identifier).
- `state` (String) - Current key state.
- `flags` (Number) - DNSKEY flags (256 for ZSK, 257 for KSK/CSK).
- `public_key` (String) - Base64-encoded public key.
- `ds_records` (List of String) - DS records for registrar (KSK/CSK only).

## Algorithm Reference

| Value | Name | Description | Recommended |
|-------|------|-------------|-------------|
| 8 | RSASHA256 | RSA with SHA-256 | ✓ |
| 10 | RSASHA512 | RSA with SHA-512 | ✓ |
| 13 | ECDSAP256SHA256 | ECDSA P-256 with SHA-256 | ✓ (Recommended) |
| 14 | ECDSAP384SHA384 | ECDSA P-384 with SHA-384 | ✓ |
| 15 | ED25519 | Edwards-curve 25519 | ✓ |
| 16 | ED448 | Edwards-curve 448 | ✓ |

### Key Size Defaults by Algorithm

| Algorithm | Default Size | Valid Range |
|-----------|-------------|-------------|
| 8 (RSASHA256) | 2048 | 1024-4096 |
| 10 (RSASHA512) | 2048 | 1024-4096 |
| 13 (ECDSAP256SHA256) | 256 | Fixed |
| 14 (ECDSAP384SHA384) | 384 | Fixed |
| 15 (ED25519) | 256 | Fixed |
| 16 (ED448) | 456 | Fixed |

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `key_tag` - The key tag used to identify this key.
- `state` - Current state of the key (e.g., `active`, `published`).
- `flags` - DNSKEY flags value.
- `public_key` - The public key in base64 format.
- `ds_records` - List of DS records to configure at your domain registrar.

## DS Record Output

The `ds_records` attribute contains DS records formatted for your registrar:

```hcl
output "ds_records_for_registrar" {
  value = bind9_dnssec_key.ksk.ds_records
}
```

Example output:
```
[
  "example.com. IN DS 12345 13 2 ABC123DEF456...",
  "example.com. IN DS 12345 13 4 789XYZ..."
]
```

## Key Rollover

To perform a key rollover, create a new key before removing the old one:

```hcl
# Existing KSK
resource "bind9_dnssec_key" "ksk_old" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 13
}

# New KSK (for rollover)
resource "bind9_dnssec_key" "ksk_new" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 15  # Upgrade to ED25519
  sign_zone = true
}

# After DS records propagate (24-48 hours), remove ksk_old
```

## Notes

### DNSSEC Best Practices

1. **Use ECDSA or EdDSA**: Algorithms 13-16 are more efficient than RSA
2. **Two-key setup**: Use separate KSK and ZSK for easier key management
3. **Key rollover**: Plan for regular ZSK rollovers (every 1-3 months)
4. **DS record propagation**: Wait 24-48 hours after updating DS records at registrar
5. **Monitor**: Check DNSSEC validation regularly

### Key Types Explained

| Type | Purpose | DS at Registrar? | Signs |
|------|---------|------------------|-------|
| KSK | Key Signing Key | Yes | DNSKEY RRset |
| ZSK | Zone Signing Key | No | All other RRsets |
| CSK | Combined Signing Key | Yes | All RRsets |

### Common Issues

1. **DS record mismatch**: Ensure DS records at registrar match output
2. **Key timing**: Allow time for DNS propagation during rollovers
3. **Algorithm downgrade**: Avoid using older algorithms (< 8)

## Import

DNSSEC keys cannot be imported as they contain sensitive cryptographic material. Create new keys using Terraform instead.

## Lifecycle

DNSSEC keys are immutable - changing any argument except `sign_zone` will force recreation of the key. Plan key rollovers carefully to avoid DNSSEC validation failures.

