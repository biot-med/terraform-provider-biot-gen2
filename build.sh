go build -o terraform-provider-biot-gen2 && \
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.1/darwin_arm64 && \
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot/1.0.1/darwin_arm64/