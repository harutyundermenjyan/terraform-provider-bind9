---
page_title: "Bulk Record Generation Guide"
subcategory: "Guides"
description: |-
  Learn how to generate multiple DNS records programmatically using Terraform's range() and for expressions.
---

# Bulk Record Generation Guide

This guide shows how to generate multiple DNS records programmatically, replacing BIND9's `$GENERATE` directive with native Terraform/OpenTofu patterns.

## BIND9 $GENERATE vs Terraform

### BIND9 $GENERATE Directive

In traditional BIND9 zone files:

```bind
; Generate host-1 to host-254 A records
$GENERATE 1-254 host-$ A 10.0.2.$

; Generate PTR records
$GENERATE 1-254 $ PTR host-$.example.com.
```

### Terraform Equivalent

In Terraform/OpenTofu, use `range()` with `for` expressions:

```terraform
locals {
  # Generate host-1 to host-254 A records
  generated_hosts = {
    for i in range(1, 255) : "host-${i}_A" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
    }
  }
}
```

## Basic Patterns

### Sequential A Records

Generate A records with sequential IPs:

```terraform
locals {
  # Creates: host-1.example.com → 10.0.2.1
  #          host-2.example.com → 10.0.2.2
  #          ...
  #          host-50.example.com → 10.0.2.50
  hosts = {
    for i in range(1, 51) : "host-${i}" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
    }
  }
}

resource "bind9_record" "hosts" {
  for_each = local.hosts

  zone    = "example.com"
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

### PTR Records (Reverse DNS)

Generate corresponding PTR records:

```terraform
locals {
  # For zone: 2.0.10.in-addr.arpa
  # Creates: 1 PTR host-1.example.com.
  #          2 PTR host-2.example.com.
  #          ...
  ptr_records = {
    for i in range(1, 51) : "${i}_PTR" => {
      name    = "${i}"
      type    = "PTR"
      ttl     = 300
      records = ["host-${i}.example.com."]
    }
  }
}

resource "bind9_record" "ptr" {
  for_each = local.ptr_records

  zone    = "2.0.10.in-addr.arpa"
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

### Custom Naming Patterns

```terraform
locals {
  # Pattern: web-01, web-02, ..., web-10
  web_servers = {
    for i in range(1, 11) : "web-${format("%02d", i)}" => {
      name    = "web-${format("%02d", i)}"  # web-01, web-02, etc.
      type    = "A"
      ttl     = 300
      records = ["10.0.10.${i}"]
    }
  }

  # Pattern: db-a, db-b, db-c (using letters)
  db_servers = {
    for idx, letter in ["a", "b", "c", "d"] : "db-${letter}" => {
      name    = "db-${letter}"
      type    = "A"
      ttl     = 300
      records = ["10.0.20.${idx + 1}"]
    }
  }
}
```

### With Step Values

BIND9: `$GENERATE 0-200/10 host-$ A 10.0.2.$`

Terraform equivalent:

```terraform
locals {
  # Every 10th number: 0, 10, 20, ..., 200
  stepped = {
    for i in range(0, 201, 10) : "host-${i}" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
    }
  }
}
```

## Advanced Patterns

### Multiple Record Types per Host

Generate A and PTR records together:

```terraform
locals {
  host_count = 50
  
  # A records
  a_records = {
    for i in range(1, local.host_count + 1) : "host-${i}_A" => {
      zone    = "example.com"
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
    }
  }
  
  # Corresponding PTR records
  ptr_records = {
    for i in range(1, local.host_count + 1) : "host-${i}_PTR" => {
      zone    = "2.0.10.in-addr.arpa"
      name    = "${i}"
      type    = "PTR"
      ttl     = 300
      records = ["host-${i}.example.com."]
    }
  }
}
```

### From CSV/List Data

Generate records from a list:

```terraform
locals {
  servers = [
    { name = "web1",    ip = "10.0.1.10" },
    { name = "web2",    ip = "10.0.1.11" },
    { name = "db1",     ip = "10.0.1.20" },
    { name = "db2",     ip = "10.0.1.21" },
    { name = "cache1",  ip = "10.0.1.30" },
  ]
  
  server_records = {
    for server in local.servers : "${server.name}_A" => {
      name    = server.name
      type    = "A"
      ttl     = 300
      records = [server.ip]
    }
  }
}

resource "bind9_record" "servers" {
  for_each = local.server_records

  zone    = "example.com"
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}
```

### From Map Data

```terraform
locals {
  # Define servers as a map
  server_ips = {
    "www"     = "10.0.1.100"
    "api"     = "10.0.1.101"
    "mail"    = "10.0.1.102"
    "ftp"     = "10.0.1.103"
    "vpn"     = "10.0.1.104"
  }
  
  server_records = {
    for name, ip in local.server_ips : "${name}_A" => {
      name    = name
      type    = "A"
      ttl     = 300
      records = [ip]
    }
  }
}
```

### Multi-Server Deployment

Generate records once, deploy to multiple servers:

```terraform
locals {
  enabled_servers = {
    for name, server in var.servers : name => server
    if server.enabled
  }

  # Define records with server targeting
  generated_records = {
    for i in range(1, 51) : "host-${i}_A" => {
      name    = "host-${i}"
      type    = "A"
      ttl     = 300
      records = ["10.0.2.${i}"]
      servers = []  # Empty = deploy to all servers
    }
  }

  # Expand to all target servers
  records_expanded = merge([
    for record_key, record in local.generated_records : {
      for server_name, server in local.enabled_servers :
      "${record_key}_${server_name}" => merge(record, { server = server_name })
      if length(record.servers) == 0 || contains(record.servers, server_name)
    }
  ]...)
}

# Create on dns1
resource "bind9_record" "generated_dns1" {
  for_each = {
    for k, v in local.records_expanded : k => v
    if v.server == "dns1"
  }
  provider = bind9.dns1

  zone    = bind9_zone.example_dns1[0].name
  name    = each.value.name
  type    = each.value.type
  ttl     = each.value.ttl
  records = each.value.records
}

# Create on dns2
resource "bind9_record" "generated_dns2" {
  for_each = {
    for k, v in local.records_expanded : k => v
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

## range() Function Reference

| Syntax | Result |
|--------|--------|
| `range(5)` | `[0, 1, 2, 3, 4]` |
| `range(1, 5)` | `[1, 2, 3, 4]` |
| `range(1, 10, 2)` | `[1, 3, 5, 7, 9]` |
| `range(10, 0, -1)` | `[10, 9, 8, 7, 6, 5, 4, 3, 2, 1]` |

## Benefits Over $GENERATE

| Feature | $GENERATE | Terraform |
|---------|-----------|-----------|
| State tracking | ❌ No | ✅ Each record tracked |
| Drift detection | ❌ No | ✅ Yes |
| Selective updates | ❌ No | ✅ Yes |
| Complex patterns | Limited | ✅ Full programming |
| Multi-server | ❌ No | ✅ Built-in |
| Version control | Zone file | ✅ HCL files |

## Tips

1. **Use meaningful keys** - The map key becomes the resource identifier
2. **Use `format()` for padding** - `format("%02d", i)` → `01`, `02`, etc.
3. **Combine with merge()** - Merge multiple generated sets together
4. **Test with `terraform console`** - Preview generated values before applying
