vault server -dev &
export VAULT_ADDR='http://127.0.0.1:8200'
sleep 1
echo -n "type in root roleid: "
read ROOT
export VAULT_TOKEN="$ROOT"
vault kv put secret/hello foo=world
vault auth enable approle
vault policy write mury vault-policy-mury.txt
vault write auth/approle/role/root policies=default,mury bind_secret_id=false token_bound_cidrs=127.0.0.1/24
vault read auth/approle/role/root/role-id
echo -n "type in roleid: "
read ROLEID
vault write auth/approle/login role_id=$ROLEID
echo "$ROLEID" > /root/.vault-roleid
vault kv list secret
vault kv get secret/hello
