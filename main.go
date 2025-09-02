// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"biot.com/terraform-provider-biot/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string = "1.0.0"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// TODO: Update address localhost to a registry after depoying the plugin publicly.
		// Address: "registry.terraform.io/localhost/biot",
		Address: "example.com/biot/biot",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		// Use log.Fatal for main function as tflog is not available here
		log.Fatal(err.Error())
	}
}
