
## Initialize Go modules (if not done)
- go mod tidy (This will clean up and download any missing dependencies)

## Build command
- go build -o terraform-provider-biot

## Run after build (to have the plugin installed locally) - (note that the version in this example is 1.0.0 and should be changed if required...)
- mkdir -p ~/.terraform.d/plugins/example.com/biot/biot/1.0.0/darwin_arm64
- cp terraform-provider-biot ~/.terraform.d/plugins/example.com/biot/biot/1.0.0/darwin_arm64/

## After updating the code - (note that the version in this example is 1.0.0 and should be changed if required...)
- go build -o terraform-provider-biot
- cp terraform-provider-biot ~/.terraform.d/plugins/example.com/biot/biot/1.0.0/darwin_arm64/