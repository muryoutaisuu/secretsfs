In this section some examples are explained on how to use specific things.

# Configuration

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

  # logging levels may be: {debug,info,warn,error}
  logging:
    level: info

  # fuse does not allow the character '/' inside of names of directories or files
  # in vault k=v pairs of one secret will be shown as files, where k is the name
  # of the file and v the value. k may also include names with a '/'.
  # Those slashes will be substituted with the following character
  # may also use some special characters, e.g. '§' or '°'
  substchar: _

# TLS Configurations
tls:
  #cacert: <path to PEM-encoded CA file>
  #capath: <path to directory of PEM-encoded CA files>
  #clientcert: <path to certificate for backend communication>
  #clientkey: <path to private key for backend communication>
  #tlsservername: <used for setting SNI host>
  #insecure: <disable TLS verification>

fio:
  enabled:
    - secretsfiles
    - templatefiles
  templatefiles:
    templatespath: /etc/secretsfs/templates/

store:
  enabled: vault
  vault:
    roleid:
      # path configuration defines, where to look for the vault roleid token
      # $HOME will be substituted with the users corresponding home directory
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

    # taken from https://www.vaultproject.io/api/secret/kv/kv-v2.html
    mtdata: secret/
    dtdata: secret/
```

# Templating

The _TemplateFilesFIO_ works with a default directory, in which templatefiles are located and which is configured in the configuration file.
Secrets will be loaded from currently active store and are called inside of the template by following string:

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

_Note: Also see the file called [`templatefile.conf`](https://github.com/muryoutaisuu/secretsfs/blob/master/example/templatefile.conf)_

# Mounting with Mountoptions

Mountoptions may be given like in a normal mount command, e.g.:

```
./secretsfs <mountpath> -o allow_other
```
