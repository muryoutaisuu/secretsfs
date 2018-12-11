export SECRETSFS_FILE_ROLEID=".vault-roleid"
export VAULT_ADDR='http://127.0.0.1:8200'
umount /mnt/secretsfs
go build ./cmd/secretsfs
./secretsfs /mnt/secretsfs -o allow_other &
