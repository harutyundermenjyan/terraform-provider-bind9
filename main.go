// BIND9 Terraform Provider
//
// A Terraform/OpenTofu provider for managing DNS zones and records
// via the BIND9 REST API.
//
// Copyright (c) 2024 Harutyun Dermenjyan
// Licensed under the MIT License
//
// Repository: https://github.com/harutyundermenjyan/terraform-provider-bind9

package main

import (
	"context"
	"flag"
	"log"

	"github.com/harutyundermenjyan/terraform-provider-bind9/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Version information (set via ldflags during build)
var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/harutyundermenjyan/bind9",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
