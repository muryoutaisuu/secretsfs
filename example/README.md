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

### FIO
ENABLED_FIOS:
- secretsfiles
- templatefiles

# templatefiles
PATH_TO_TEMPLATES: /etc/secretsfs/templates/

### STORE
CURRENT_STORE: Vault

# vault
FILE_ROLEID: .vault-roleid
VAULT_ADDR: http://127.0.0.1:8200
# taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
MTDATA: secret/metadata/
DTDATA: secret/data/

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

### fstab Configuration

Does not yet work.
Problem is, that _secretsfs_ isn't yet daemonized, and hence the tool will run without returning the prompt, which is problematic for fstab mounts.

Nevertheless, a working fstab configuration would look like:

```fstab
secretsfs       /mnt/fstabsecretsfs     fuse    allow_other     0 0
```

_Note: Also see the file called fstab_
