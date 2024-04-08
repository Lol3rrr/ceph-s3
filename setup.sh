#!/bin/bash

set +xe;

SHA256=$(shasum -a 256 vault/plugins/ceph-s3 | awk 'BEGIN {FS=" ";} {print $1}');
echo "SHA256: '$SHA256'";

vault write sys/plugins/catalog/database/test-db sha256="$SHA256" command="ceph-s3";

vault secrets enable database;
vault write database/config/test-database plugin_name=test-db allowed_roles="*" ceph_username="$CEPH_USERNAME" ceph_password="$CEPH_PASSWORD" ceph_url="$CEPH_URL";

vault write database/roles/test-role db_name=test-database default_ttl="1h" max_ttl="24h" creation_statements="testing" revocation_statements="testing"
