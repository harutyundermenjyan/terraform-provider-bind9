# BIND9 Terraform Provider - Example Usage

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
  endpoint = var.bind9_endpoint  # e.g., "https://dns.example.com:8080"
  api_key  = var.bind9_api_key   # API key authentication
  
  # Or use username/password:
  # username = var.bind9_username
  # password = var.bind9_password
  
  # Optional settings
  insecure = false  # Skip TLS verification
  timeout  = 30     # Request timeout in seconds
}

# Variables
variable "bind9_endpoint" {
  description = "BIND9 REST API endpoint"
  type        = string
  default     = "http://localhost:8080"
}

variable "bind9_api_key" {
  description = "BIND9 API key"
  type        = string
  sensitive   = true
}

# =============================================================================
# Access Control Lists (ACLs)
# =============================================================================

# Internal networks ACL
resource "bind9_acl" "internal" {
  name = "internal"
  
  entries = [
    "localhost",
    "localnets",
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
  ]
  
  comment = "RFC1918 private networks"
}

# Secondary DNS servers allowed for zone transfers
resource "bind9_acl" "secondaries" {
  name = "secondaries"
  
  entries = [
    "192.168.1.11",           # ns2
    "192.168.1.12",           # ns3
    "key \"transfer-key\"",   # TSIG authenticated transfers
  ]
  
  comment = "Secondary DNS servers for zone transfers"
}

# Dynamic DNS update clients
resource "bind9_acl" "ddns_clients" {
  name = "ddns-clients"
  
  entries = [
    "192.168.2.0/24",         # DHCP server network
    "key \"ddns-key\"",       # TSIG key for updates
  ]
  
  comment = "Hosts allowed for dynamic DNS updates"
}

# =============================================================================
# Zone Management
# =============================================================================

# Create a master zone using ACLs for access control
resource "bind9_zone" "example" {
  name = "example.com"
  type = "master"
  
  # SOA record configuration
  soa_mname   = "ns1.example.com"
  soa_rname   = "hostmaster.example.com"
  soa_refresh = 86400
  soa_retry   = 7200
  soa_expire  = 3600000
  soa_minimum = 3600
  default_ttl = 3600
  
  # Initial nameservers with glue records
  nameservers = [
    "ns1.example.com",
    "ns2.example.com",
  ]
  
  ns_addresses = {
    "ns1.example.com" = "192.168.1.10"
    "ns2.example.com" = "192.168.1.11"
  }
  
  # Access control using ACLs defined above
  allow_query    = ["internal", "any"]    # Internal + public queries
  allow_transfer = ["secondaries"]         # Only to secondary NS
  allow_update   = ["ddns-clients"]        # Only DDNS clients
  
  notify = true
  
  # Delete zone file when zone is destroyed
  delete_file_on_destroy = false

  # Ensure ACLs exist before creating zone
  depends_on = [
    bind9_acl.internal,
    bind9_acl.secondaries,
    bind9_acl.ddns_clients,
  ]
}

# =============================================================================
# A Records
# =============================================================================

resource "bind9_record" "ns1" {
  zone    = bind9_zone.example.name
  name    = "ns1"
  type    = "A"
  ttl     = 3600
  records = ["192.168.1.10"]
}

resource "bind9_record" "ns2" {
  zone    = bind9_zone.example.name
  name    = "ns2"
  type    = "A"
  ttl     = 3600
  records = ["192.168.1.11"]
}

resource "bind9_record" "www" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "A"
  ttl     = 300
  records = ["192.168.1.100"]
}

# Multiple A records for the same name (round-robin)
resource "bind9_record" "app" {
  zone    = bind9_zone.example.name
  name    = "app"
  type    = "A"
  ttl     = 60
  records = [
    "192.168.1.101",
    "192.168.1.102",
    "192.168.1.103",
  ]
}

# =============================================================================
# AAAA Records (IPv6)
# =============================================================================

resource "bind9_record" "www_ipv6" {
  zone    = bind9_zone.example.name
  name    = "www"
  type    = "AAAA"
  ttl     = 300
  records = ["2001:db8::100"]
}

# =============================================================================
# CNAME Records
# =============================================================================

resource "bind9_record" "blog" {
  zone    = bind9_zone.example.name
  name    = "blog"
  type    = "CNAME"
  ttl     = 3600
  records = ["www.example.com."]
}

resource "bind9_record" "shop" {
  zone    = bind9_zone.example.name
  name    = "shop"
  type    = "CNAME"
  ttl     = 3600
  records = ["shopify.example.com."]
}

# =============================================================================
# MX Records (Mail)
# =============================================================================

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

# =============================================================================
# TXT Records
# =============================================================================

# SPF record
resource "bind9_record" "spf" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["v=spf1 include:_spf.google.com ~all"]
}

# DKIM record
resource "bind9_record" "dkim" {
  zone    = bind9_zone.example.name
  name    = "google._domainkey"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4..."]
}

# DMARC record
resource "bind9_record" "dmarc" {
  zone    = bind9_zone.example.name
  name    = "_dmarc"
  type    = "TXT"
  ttl     = 3600
  records = ["v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com"]
}

# Verification TXT record
resource "bind9_record" "verification" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "TXT"
  ttl     = 3600
  records = ["google-site-verification=abc123xyz"]
}

# =============================================================================
# SRV Records
# =============================================================================

# SIP service
resource "bind9_record" "sip_tcp" {
  zone    = bind9_zone.example.name
  name    = "_sip._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["10 60 5060 sip.example.com."]
}

# XMPP service
resource "bind9_record" "xmpp" {
  zone    = bind9_zone.example.name
  name    = "_xmpp-server._tcp"
  type    = "SRV"
  ttl     = 3600
  records = ["10 0 5269 xmpp.example.com."]
}

# =============================================================================
# CAA Records (Certificate Authority Authorization)
# =============================================================================

resource "bind9_record" "caa" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "CAA"
  ttl     = 3600
  records = [
    "0 issue \"letsencrypt.org\"",
    "0 issuewild \"letsencrypt.org\"",
    "0 iodef \"mailto:security@example.com\"",
  ]
}

# =============================================================================
# NS Records (Additional)
# =============================================================================

resource "bind9_record" "ns" {
  zone    = bind9_zone.example.name
  name    = "@"
  type    = "NS"
  ttl     = 86400
  records = [
    "ns1.example.com.",
    "ns2.example.com.",
  ]
}

# =============================================================================
# PTR Records (Reverse DNS)
# =============================================================================

# Note: You would typically create a reverse zone for this
# resource "bind9_zone" "reverse" {
#   name = "1.168.192.in-addr.arpa"
#   type = "master"
# }
# 
# resource "bind9_record" "ptr_www" {
#   zone    = bind9_zone.reverse.name
#   name    = "100"
#   type    = "PTR"
#   ttl     = 3600
#   records = ["www.example.com."]
# }

# =============================================================================
# DNSSEC
# =============================================================================

# Generate KSK (Key Signing Key)
resource "bind9_dnssec_key" "ksk" {
  zone      = bind9_zone.example.name
  key_type  = "KSK"
  algorithm = 13  # ECDSAP256SHA256
  sign_zone = true
}

# Generate ZSK (Zone Signing Key)
resource "bind9_dnssec_key" "zsk" {
  zone      = bind9_zone.example.name
  key_type  = "ZSK"
  algorithm = 13  # ECDSAP256SHA256
  
  depends_on = [bind9_dnssec_key.ksk]
}

# =============================================================================
# Data Sources
# =============================================================================

# Get zone information
data "bind9_zone" "info" {
  name = bind9_zone.example.name
}

# List all zones
data "bind9_zones" "all" {}

# Get specific record
data "bind9_record" "www_lookup" {
  zone = bind9_zone.example.name
  name = "www"
  type = "A"
  
  depends_on = [bind9_record.www]
}

# List all A records in zone
data "bind9_records" "a_records" {
  zone = bind9_zone.example.name
  type = "A"
  
  depends_on = [bind9_record.www, bind9_record.app]
}

# =============================================================================
# Outputs
# =============================================================================

output "zone_serial" {
  description = "Current zone serial number"
  value       = data.bind9_zone.info.serial
}

output "zone_count" {
  description = "Total number of zones"
  value       = length(data.bind9_zones.all.zones)
}

output "www_ips" {
  description = "IP addresses for www"
  value       = data.bind9_record.www_lookup.records
}

output "dnssec_ds_records" {
  description = "DS records for registrar (DNSSEC)"
  value       = bind9_dnssec_key.ksk.ds_records
}

output "a_record_count" {
  description = "Number of A records in zone"
  value       = length(data.bind9_records.a_records.records)
}

output "acls_configured" {
  description = "ACLs configured for this zone"
  value = {
    internal    = bind9_acl.internal.name
    secondaries = bind9_acl.secondaries.name
    ddns        = bind9_acl.ddns_clients.name
  }
}