# Configuration File

The default configuration of secretsfs may be output with the command `./secretsfs --print-defaults` and returns the output described below.
This output may be directly piped into the actual configuration file: `./secretsfs --print-defaults > /etc/secretsfs/secretsfs.yaml`.
It's best to generate a new configuration file from the binary and then reconfigure it to one's needs.

```yaml
---
# General
general:
  configuration:
    paths:
      #- /etc/secretsfs/
      #- $HOME/.secretsfs
    #configfile: secretsfs  # without file type

  # logging levels may be: {trace,debug,info,warn,error,fatal,panic}
  logging:
    level: info

fio:
  enabled:
    - secretsfiles
    - templatefiles
    - internal
  templatefiles:
    # add additional locations for template files
    # the files in '/etc/secretsfs/templates/' for example will be mapped to
    # 'templatefiles/default/'
    templatespaths:
      default: /etc/secretsfs/templates/
      #applA: /appl/applA
  secretsfiles:
  internal:
    # privileges given to users or groups for listing and reading files in internal
    # do not make this readable for all, as it may contain critical data due to path namings
    privileges:
      users:
        - root
      groups:
        - admin

store:
  enabled: vault
  vault:
    roleid:
      # path configuration defines, where to look for the vault roleid token
      # $HOME will be substituted with the user's corresponding home directory
      # according to variable HomeDir in https://golang.org/pkg/os/user/#User
      # it *MUST* be uppcerase
      file: "$HOME/.vault-roleid"

      # useroverride configures paths per user, may be used to overwrite default
      # store.vault.roleid.file for some users
      # takes precedence over store.vault.roleid.file
      # store.vault.roleid.useroverride will *NOT* fallback to store.vault.roleid.file
      #useroverride:
      #  <usernameA>: <path>

    # address of the vault instance, that shall be accessed
    # differenciates between http:// and https:// protocols
    # defaults to a local dev instance
    addr: http://127.0.0.1:8200

    # vault TLS Configurations
    # for more information, see https://pkg.go.dev/github.com/hashicorp/vault/api#TLSConfig
    tls:
      #cacert: <path to PEM-encoded CA file>
      #capath: <path to directory of PEM-encoded CA files>
      #clientcert: <path to certificate for backend communication>
      #clientkey: <path to private key for backend communication>
      #tlsservername: <used for setting SNI host>
      #insecure: <disable TLS verification>
```

# Templating

The _TemplateFiles FIO_ works with configurable directories in which templatefiles are placed as needed.
The directories are configurable in the configuration file via `fio.templatefiles.templatespaths` and map paths from _secretsfs_ to other filesystems.
The path of the secret in Vault can be copied from the _SecretsFiles FIO_ and includes the name of the used key of the secret.
If the calling user has no permissions in vault to access at least one of the secrets in the templatefile, _TemplateFiles FIO_ will return an error.
Secrets will be loaded from currently active store and are called inside of the template by following string:

```
{{ .Get "<pathToSecret>" }}
```

*__Note:__ The Quotes '"' around the `<pathToSecret>` are very important.
That is due to golang templating notation, so that the input will be validated as a string.
If it is not quoted, the golang templating library will not validate the input as a string and hence secretsfs will return an error.*

Any (non-)standard textfile configuration file formats may be used.
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

_Note: Also see the file called [`example/templatefile.conf`](https://github.com/muryoutaisuu/secretsfs/blob/master/example/templatefile.conf)_

# Mounting with Mountoptions

Mountoptions may be given like in a normal mount command, e.g.:

```
./secretsfs <mountpath> -o allow_other
```
