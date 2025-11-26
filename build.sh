go build -o terraform-provider-biot-gen2 && \
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/1.0.2/darwin_arm64 && \
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/1.0.2/darwin_arm64/