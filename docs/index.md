# secretsfs

**TL;DR**: Access your secrets securely via a simple `cat` command instead of using a client.

_secretsfs_ implements a [FUSE](https://en.wikipedia.org/wiki/Filesystem_in_Userspace)-filesytem, that allows you to interact with secrets stored in a backend (called store) via simple readonly filesysten-interacting commands, like cat, grep etc.
One such store may be [Vault](https://github.com/hashicorp/vault).

Output formats _File Input/Output_ (FIO) are treated like plugins and can be (de-)activated in a configuration file. Out of the box implemented FIOs are:

* **secretsfiles:** returns plain secret on a simple `cat`
* **templatefiles:** returns on `cat` a with secrets rendered file (e.g. a configuration file with secrets)

Get it now on [GitHub](https://github.com/muryoutaisuu/secretsfs)!


# Getting started

## Just Test _secretsfs_

If you simply want to test how to interact with _secretsfs_, you may execute the shell script `vault-startup.sh`.
This just starts a development Vault  instance on your host listening on port 8200 and populates some initial values.
It also makes sure, that your root user has the correct roleid value set in his Vault  access file.

```bash
mkdir /mnt/secretsfs                                # create the default mountpoint
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs    # change to directory
./vault-setup.sh                                    # setup Vault, just following instructions on screen
                                                    # yes, I was too lazy to do some string parsing
. sourceit                                          # source some environment variables
./clean.sh                                          # this umounts potentially existing old mounts, build secretsfs anew and mounts it
                                                    # Type <ENTER> so you can see your prompt again
ls /mnt/secretsfs/secretsfiles                      # look at entries inside of that new secretsfs
```

If you also want to see templatefiles in action, additional actions are need:

```bash
mkdir -p /etc/secretsfs/templatefiles
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs            # change to directory
cp examples/templatefile.conf /etc/secretsfs/templatefiles  # copy the template example to the templatefiles
ls /mnt/secretsfs/templatefiles                             # see that the newly copied file now gets listed
cat /mnt/secretsfs/templatefiles/templatefile.conf          # see that the secret is rendered upon this cat
```

## Install _secretsfs_

### From Source

To get started with _secretsfs_ simply download this repository and install it:

```bash
go get github.com/muryoutaisuu/secretsfs            # get main source
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs    # change to directory
go get ./...                                        # get dependencies
go install ./cmd/secretsfs                          # install secretsfs
```

### Just Copy the Binary

For Linux x86_64, you may download the latest built binary from [the project's release page](https://github.com/muryoutaisuu/secretsfs/releases).

### RPM Package

For RHEL7, there is a prebuilt .rpm package, you may download from [the project's release page](https://github.com/muryoutaisuu/secretsfs/releases).

## Start *secretsfs*

There are two possible ways to start *secretsfs*:

* Start it manually with `secretsfs <mountpath> [-o <mountoptions>] [&]]`
* Start it with Systemd, use the predefined service in the examples folder
** `cp example/secretsfs.service /usr/lib/systemd/system/secretsfs.service`
** `systemctl start secretsfs`
** `systemctl enable secretsfs`


# Examples

In this section some examples are explained on how to use specific things.

## Configuration

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
# $HOME will be substituted with the users corresponding home directory
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

## Templating

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

## Mounting with Mountoptions

Mountoptions may be given like in a normal mount command, e.g.:

```
./secretsfs <mountpath> -o allow_other
```


# Known Issues

* **Substitution:** `a/b` may be substituted to `a_b`, which may also already be in the backend (e.g. Vault). This will likely cause a clash. As a workaround either configure `subst_char` in the configuration file to a different value, or do not use `/` in Vault key names at all. If clashing, the alphabetically first key name will have precedence (in this case it would be `a/b`).
* In Vault, both paths `/secret/foo` and `/secret/foo/` may exist, where the former is a secret and the latter is a subpath. Filesystems know no difference between a path with and one without the `/` at the end. Hence Both validate to the same path. In _secretsfs_ this results into the keys of `/secret/foo` being displayed as files next to the subdirectory `/secret/foo/`, while in reality those two are not connected in any way to each other in Vault.
