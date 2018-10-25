vault server -dev &
export VAULT_ADDR='http://127.0.0.1:8200'
sleep 1
echo -n "type in root token: "
read ROOT
export VAULT_TOKEN="$ROOT"
echo "export ROOTTOKEN=$VAULT_TOKEN" > sourceit
vault kv put secret/hello foo=world
vault kv put secret/subdir/mury foo2=world2 bar2=natii
vault auth enable approle
vault policy write mury vault-policy-mury.txt
vault write auth/approle/role/root policies=default,mury bind_secret_id=false token_bound_cidrs=127.0.0.1/24
vault read auth/approle/role/root/role-id
echo -n "type in roleid: "
read ROLEID
export ROLEID
echo "$ROLEID" > /root/.vault-roleid
echo "export ROLEID=$ROLEID" >> sourceit
vault write auth/approle/login role_id=$ROLEID
echo -n "type in root approle token:"
read ROLETOKEN
export ROLETOKEN
echo "export ROLETOKEN=$ROLETOKEN" >> sourceit
echo "export VAULT_TOKEN=$ROLETOKEN" >> sourceit
vault kv list secret
vault kv get secret/hello
vault kv get secret/subdir/mury
