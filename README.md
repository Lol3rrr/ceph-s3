# Ceph-S3
A Vault plugin to allow for dynamic secret generation for ceph's rgw

## Testing
Running the vault instance
```
go build -o vault/plugins/ceph-s3 main.go
vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins -log-level=debug
```

Configuration
```
CEPH_USERNAME=xxx CEPH_PASSWORD=yyy bash setup.sh
vault read database/creds/test-role
```
