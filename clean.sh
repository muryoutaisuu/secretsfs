export SECRETSFS_FILE_ROLEID=".vault-roleid"
umount /mnt/secretsfs
go build ./cmd/secretsfs
./secretsfs /mnt/secretsfs &
