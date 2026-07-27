VERSION=$(grep 'version string' main.go | grep -o '"[^"]*"' | tr -d '"')

go build -o terraform-provider-biot-gen2 && \
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/${VERSION}/darwin_arm64 && \
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/${VERSION}/darwin_amd64 && \
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/${VERSION}/darwin_arm64/ && \
cp terraform-provider-biot-gen2 ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/${VERSION}/darwin_amd64/
