---
page_title: "bind9_dnssec_key Resource - BIND9 Provider"
subcategory: "DNSSEC"
description: |-
  Manages a DNSSEC key for a DNS zone.
---

# bind9_dnssec_key (Resource)

Manages DNSSEC keys for DNS zones on a BIND9 server. Supports generating Key Signing Keys (KSK), Zone Signing Keys (ZSK), and Combined Signing Keys (CSK) with various cryptographic algorithms.

## Example Usage

### Generate KSK (Key Signing Key)

```terraform
resource "bind9_dnssec_key" "ksk" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}
```

### Generate ZSK (Zone Signing Key)

```terraform
resource "bind9_dnssec_key" "zsk" {
  zone      = "example.com"
  key_type  = "ZSK"
  algorithm = 13  # ECDSAP256SHA256
}
```

### Generate CSK (Combined Signing Key)

```terraform
resource "bind9_dnssec_key" "csk" {
  zone      = "example.com"
  key_type  = "CSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}
```

### RSA Key with Custom Size

```terraform
resource "bind9_dnssec_key" "rsa_ksk" {
  zone      = "example.com"
  key_type  = "KSK"
  algorithm = 8     # RSASHA256
  bits      = 2048  # Key size in bits
  sign_zone = true
}
```

### Complete DNSSEC Setup

```terraform
# First, create the zone
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"

  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  nameservers = ["ns1.example.com", "ns2.example.com"]
}

# Generate KSK (Key Signing Key)
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.example.name
  key_type  = "KSK"
  algorithm = 13
  ttl       = 3600
}

# Generate ZSK (Zone Signing Key)
resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.example.name
  key_type  = "ZSK"
  algorithm = 13
  ttl       = 300
  sign_zone = true  # Sign zone after creating ZSK
}

# Output DS records for registrar
output "ds_records" {
  value       = bind9_dnssec_key.ksk.ds_records
  description = "DS records to submit to domain registrar"
}
```

## Argument Reference

### Required

- `zone` (String) The zone name to create the DNSSEC key for. **Changing this forces a new resource to be created.**
- `key_type` (String) The type of DNSSEC key. **Changing this forces a new resource to be created.** Valid values:
  - `KSK` - Key Signing Key (signs DNSKEY records)
  - `ZSK` - Zone Signing Key (signs zone data)
  - `CSK` - Combined Signing Key (acts as both KSK and ZSK)

### Optional

- `algorithm` (Number) DNSSEC algorithm number. Default: `13` (ECDSAP256SHA256). See algorithm reference below.
- `bits` (Number) Key size in bits. Only applicable to RSA algorithms. ECDSA and EdDSA algorithms have fixed key sizes.
- `ttl` (Number) TTL for the DNSKEY record. Default: `3600` (1 hour)
- `sign_zone` (Boolean) Whether to sign the zone after creating the key. Set to `true` on the last key creation to trigger zone signing.

### Read-Only

- `id` (String) The key identifier in format `zone/key_tag`.
- `key_tag` (Number) The DNSKEY key tag (computed). Used to identify the key.
- `state` (String) Current key state (e.g., `ACTIVE`, `PUBLISHED`, `RETIRED`).
- `flags` (Number) DNSKEY flags value (256 for ZSK, 257 for KSK/CSK).
- `public_key` (String) Base64-encoded public key data.
- `ds_records` (List of String) DS (Delegation Signer) records for this key. Submit these to your domain registrar for the chain of trust.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The key identifier.
- `key_tag` - The DNSKEY key tag.
- `state` - The current key state.
- `flags` - The DNSKEY flags.
- `public_key` - The base64-encoded public key.
- `ds_records` - The DS records for registrar submission.

## Algorithm Reference

| Value | Algorithm | Key Size | Notes |
|-------|-----------|----------|-------|
| 8 | RSASHA256 | 1024-4096 bits | RSA with SHA-256. Widely supported. |
| 10 | RSASHA512 | 1024-4096 bits | RSA with SHA-512. |
| 13 | ECDSAP256SHA256 | 256 bits (fixed) | **Recommended.** Smaller keys, good performance. |
| 14 | ECDSAP384SHA384 | 384 bits (fixed) | Higher security than P-256. |
| 15 | ED25519 | 256 bits (fixed) | Modern, fast, secure. Growing support. |
| 16 | ED448 | 448 bits (fixed) | Highest security. Limited support. |

### Algorithm Recommendations

| Use Case | Recommended Algorithm | Reason |
|----------|----------------------|--------|
| General use | 13 (ECDSAP256SHA256) | Best balance of security, performance, and compatibility |
| Maximum compatibility | 8 (RSASHA256) | Supported by all resolvers |
| Maximum security | 14 or 16 | Higher security margins |
| Modern infrastructure | 15 (ED25519) | Fastest, modern crypto |

## Key Types Explained

### KSK (Key Signing Key)

- **Purpose:** Signs the DNSKEY RRset (the zone's public keys)
- **Flags:** 257 (SEP bit set)
- **Lifecycle:** Changed infrequently (yearly or less)
- **DS Record:** Must be submitted to parent zone/registrar
- **Size:** Can be larger for added security

### ZSK (Zone Signing Key)

- **Purpose:** Signs all other records in the zone
- **Flags:** 256
- **Lifecycle:** Rotated more frequently (monthly to quarterly)
- **DS Record:** Not submitted to registrar
- **Size:** Smaller for better performance (signs many records)

### CSK (Combined Signing Key)

- **Purpose:** Acts as both KSK and ZSK
- **Flags:** 257
- **Lifecycle:** Simplified key management
- **DS Record:** Must be submitted to registrar
- **Use Case:** Simpler setup, suitable for smaller zones

## DNSSEC Workflow

### Initial Setup

1. Create the zone
2. Create KSK (or CSK)
3. Create ZSK (if using separate keys)
4. Sign the zone (set `sign_zone = true` on last key)
5. Submit DS records to registrar

### Key Rollover (ZSK)

1. Create new ZSK with `sign_zone = true`
2. Wait for TTL propagation
3. Remove old ZSK resource

### Key Rollover (KSK)

1. Create new KSK (don't remove old one yet)
2. Submit new DS record to registrar
3. Wait for DS propagation (can take 24-48 hours)
4. Remove old KSK resource
5. Remove old DS record from registrar

## Best Practices

1. **Use ECDSA (algorithm 13)** - Smaller signatures mean faster DNS responses and less bandwidth.

2. **Separate KSK and ZSK** - Allows different rotation schedules and security levels.

3. **ZSK TTL should be short** - Allows faster key rollover.

4. **KSK TTL can be longer** - KSK changes are less frequent.

5. **Monitor DS records** - Ensure DS records are correctly published at registrar.

6. **Plan for rollovers** - Document your key rotation process.

7. **Keep backup of keys** - BIND9 stores keys in the keys directory.

## Notes

- DNSSEC keys are immutable once created. To change algorithm or key type, create a new key and retire the old one.
- Zone signing may take time for large zones.
- Ensure your BIND9 server is configured to support DNSSEC operations.
- DS record propagation to the parent zone may take up to 48 hours depending on registrar.
