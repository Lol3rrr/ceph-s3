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

## Configuration
### Database Config
Option | Description
-------|------------
`ceph_username` | The Username for the Ceph User to use
`ceph_password` | The Password for the Ceph User to use
`ceph_url` | The base URL for Ceph

### Database Role
Option | Desired Value
-------|--------------
Creation Statements | The RGW User to create S3 credentials for
Revokation Statements | The RGW User to create S3 credentials for
