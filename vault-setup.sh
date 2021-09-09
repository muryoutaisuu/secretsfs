# kill existing vault
echo "killing old vault instance..."
pkill vault
sleep 1

# start vault dev server & set ENV variables
export VAULT_DEV_ROOT_TOKEN_ID="root"
vault server -dev &
export VAULT_ADDR='http://127.0.0.1:8200'
echo export VAULT_ADDR='http://127.0.0.1:8200' > sourceit

# read root token & set ENV variables
export ROOT="${VAULT_DEV_ROOT_TOKEN_ID}"
export VAULT_TOKEN="$ROOT"
#echo export VAULT_TOKEN="$ROOT" >> sourceit
echo "export ROOTTOKEN=$VAULT_TOKEN" >> sourceit

# put some secrets into vault
vault kv put secret/myappl/hello foo=world
vault kv put secret/myappl/subdir/mury foo2=world2 bar2=natii
vault kv put secret/myappl/hello2 my/bad/key=my/bad/value my_bad_key=my_bad_value mynormalkey=mynormalvalue "my key"="my value" "my second key"="my second value"
vault kv put secret/private/mysecret mykey=myvalue
vault kv put secret/private/notalloweddir/notallowedsubdir/notallowedsecret mynotallowedkey=mynotallowedsecret

# create a new approle called mury and allow it nearly everything
vault auth enable approle
vault policy write mury vault-policy-mury.txt
vault write auth/approle/role/root policies=default,mury bind_secret_id=false token_bound_cidrs=127.0.0.1/24

# get a approleid from approle mury & write it into corresponding files
vault read auth/approle/role/root/role-id
export ROLEID=$(vault read -format=json auth/approle/role/root/role-id | jq -r '.data.role_id')
#echo "$ROLEID" > /root/.vault-roleid
echo "$ROLEID" > /home/fiorettin/.vault-roleid
echo "export ROLEID=$ROLEID" >> sourceit

# login with roleid to check whether everything works as intended and also save accesstoken as roletoken
# it's rather for showing, that get/put works with non-root approles!
login=$(vault write -format=json auth/approle/login role_id=$ROLEID)
echo "$login"
export ROLETOKEN=$(echo "$login" | jq -r '.auth.client_token')
echo "export ROLETOKEN=$ROLETOKEN" >> sourceit
echo "export VAULT_TOKEN=$ROLETOKEN" >> sourceit
vault kv list secret
vault kv get secret/myappl/hello
vault kv get secret/myappl/subdir/mury
echo "now run:"
echo "  . sourceit"
echo "  echo \$ROLEID > /export/home/fiorettin/.vault-roleid"
echo "  umount /mnt/secretsfs ; go build ./cmd/secretsfs/ && ./secretsfs /mnt/secretsfs -o allow_other > /tmp/secretsfs.log"
