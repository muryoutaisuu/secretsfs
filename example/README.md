## Examples

In this section some examples on how to use specific things are explained.

### Configuration

The default configuration of secretsfs may be output with the command `./secretsfs --print-defaults` and returns output:

```yaml
---
### GENERAL
# CONFIG_PATHS:
# - /etc/secretsfs/
# - $HOME/.secretsfs
# CONFIG_FILE: secretsfs  # without file type

# HTTPS Configurations
# HTTPS_CACERT: <path to PEM-encoded CA file>
# HTTPS_CAPATH: <path to directory of PEM-encoded CA files>
# HTTPS_CLIENTCERT: <path to certificate for backend communication>
# HTTPS_CLIENTKEY: <path to private key for backend communication>
# HTTPS_TLSSERVERNAME: <used for setting SNI host>
# HTTPS_INSECURE: <disable TLS verification>


### FIO
ENABLED_FIOS:
- secretsfiles
- templatefiles

# templatefiles
PATH_TO_TEMPLATES: /etc/secretsfs/templates/


### STORE
CURRENT_STORE: Vault

# vault
# path configuration defines, where to look for the vault roleid token
# $HOME will be substituted with the user's corresponding home directory
# according to variable HomeDir in https://golang.org/pkg/os/user/#User
# old: FILE_ROLEID: .vault-roleid
FILE_ROLEID: "$HOME/.vault-roleid"

# FILE_ROLEID_USER configures paths per user, may be used to overwrite default
# FILE_ROLEID for some users
# takes precedence over FILE_ROLEID
# FILE_ROLEID_USER will *NOT* fallback to FILE_ROLEID
#FILE_ROLEID_USER:
#  <usernameA>: <path>
VAULT_ADDR: http://127.0.0.1:8200
# taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
MTDATA: secret/
DTDATA: secret/


# fuse does not allow the character '/' inside of names of directories or files
# in vault k=v pairs of one secret will be shown as files, where k is the name
# of the file and v the value. k may also include names with a '/'.
# Those slashes will be substituted with the following character
# may also use some special characters, e.g. '§' or '°'
subst_char: _
```

### Templating

The TemplateFilesFIO works with a default directory, in which templatefiles are located and which is configured in the configuration file.
Secrets will be loaded from currently activated store and are called inside of the template by following string:

```
{{ .Get "<pathToSecret>" }}
```

*__Note:__ The Quotes '"' around the pathToSecret are very important.
That is due to golang templating notation, because the input will be validated as a string.
If it is not quoted, the golang templating library will not validate the input as a string and hence secretsfs will return an error.*

Any standard textfile configuration file formats may be used.
Just to list some of the mostly spread:
* XML
* JSON
* YAML
* TOML
* any other textfile format

A complete example may look like this:

```toml
[defaults]
foo = {{ .Get "subdir/bar" }}
```

_Note: Also see the file called templatefile.conf_

### Mounting with Mountoptions

Mountoptions may be given like in a normal mount command, e.g.:

```
./secretsfs <mountpath> -o allow_other
```
