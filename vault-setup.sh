# kill existing vault
pkill vault

# start vault dev server & set ENV variables
vault server -dev &
export VAULT_ADDR='http://127.0.0.1:8200'
echo export VAULT_ADDR='http://127.0.0.1:8200' > sourceit

# read root token & set ENV variables
sleep 1
echo -n "type in root token: "
read ROOT
export VAULT_TOKEN="$ROOT"
#echo export VAULT_TOKEN="$ROOT" >> sourceit
echo "export ROOTTOKEN=$VAULT_TOKEN" >> sourceit

# put some secrets into vault
vault kv put secret/hello foo=world
vault kv put secret/subdir/mury foo2=world2 bar2=natii
vault kv put secret/hello2 my/bad/key=my/bad/value my_bad_key=my_bad_value mynormalkey=mynormalvalue "my key"="my value" "my second key"="my second value"

# create a new approle called mury and allow it nearly everything
vault auth enable approle
vault policy write mury vault-policy-mury.txt
vault write auth/approle/role/root policies=default,mury bind_secret_id=false token_bound_cidrs=127.0.0.1/24

# get a approleid from approle mury & write it into corresponding files
vault read auth/approle/role/root/role-id
echo -n "type in roleid: "
read ROLEID
export ROLEID
#echo "$ROLEID" > /root/.vault-roleid
echo "$ROLEID" > /home/fiorettin/.vault-roleid
echo "export ROLEID=$ROLEID" >> sourceit

# login with roleid to check whether everything works as intended and also save accesstoken as roletoken
# it's rather for showing, that get/put works with non-root approles!
vault write auth/approle/login role_id=$ROLEID
echo -n "type in root approle token:"
read ROLETOKEN
export ROLETOKEN
echo "export ROLETOKEN=$ROLETOKEN" >> sourceit
echo "export VAULT_TOKEN=$ROLETOKEN" >> sourceit
vault kv list secret
vault kv get secret/hello
vault kv get secret/subdir/mury
