go build -o terraform-provider-biot && \
mkdir -p ~/.terraform.d/plugins/example.com/biot/biot/1.0.0/darwin_arm64 && \
cp terraform-provider-biot ~/.terraform.d/plugins/example.com/biot/biot/1.0.0/darwin_arm64/