[![GoDoc](https://godoc.org/github.com/muryoutaisuu/secretsfs?status.svg)](https://godoc.org/github.com/muryoutaisuu/secretsfs/pkg)

* [GoDoc](https://godoc.org/github.com/muryoutaisuu/secretsfs/pkg)
* [GoWalker](https://gowalker.org/github.com/muryoutaisuu/secretsfs/pkg)

# Work in Progress...

Subject to changes of license, ownership, api and code structure amongst other things...

**Do not use yet!!!**

# Purpose of secretsfs

_secretsfs_ implements a fuse-filesytem, that allows to interact with secrets stored in a backend (called store) via simple filesysten-interacting commands.
One such store may be [Vault](https://github.com/hashicorp/vault).

Output formats (called FIO, stands for File Input/Output) are treated like plugins and can be (de-)activated in a configuration file. Out of the box implemented FIOs are:

* **secretsfiles:** returns plain secret on a simple `cat`
* **templatefiles:** returns on `cat` a with secrets rendered file (e.g. a configuration file with secrets)

# Getting started

## Just test _secretsfs_ with a dev Vault instance

If you simply want to test how to interact with _secretsfs_, you may execute the shell script `vault-startup.sh`.
This just starts a development vault instance on your host listening on port 8200 and populates some initial values.
It also makes sure, that your root user has the correct roleid value set in his vault access file.

```bash
mkdir /mnt/secretsfs                                # create the default mountpoint
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs    # change to directory
./vault-setup.sh                                    # setup vault, just following instructions on screen
                                                    # yes, I was too lazy to do some string parsing
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

To get started with _secretsfs_ simply download this repository and install it:

```bash
go get github.com/muryoutaisuu/secretsfs            # get main source
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs    # change to directory
go get ./...                                        # get dependencies
go install ./cmd/secretsfs                          # install secretsfs
```

## How to start *secretsfs*

* Start it manually with `secretsfs <mountpath> [-o <mountoptions>] [-foreground [&]]
* Start it with Systemd, use the predefined service in the examples folder
* Start it with fstab, use the predefined line in the examples folder

# Known Issues

* **Substitution:** `a/b` may be substituted to `a_b`, which may also already be in the backend (e.g. Vault). This will likely cause a clash. As a workaround either configure `subst_char` in the configuration file to a different value, or do not use `/` in Vault key names at all. If clashing, the alphabetically first key name will have precedence (in this case it would be `a/b`).
* **Background use:** if *secretsfs* is used in background, it will start itself with the `-foreground` parameter. This causes the process in `ps -ef` to be shown with the `-foreground` flag although the user started it without the -foreground flag. It's a rather aestethic issue.

# Varia

## Local GoDoc

Generate local GoDoc:

```
godoc -http :8080 &
```

Go to: `http://localhost:8080/pkg/github.com/muryoutaisuu/secretsfs`
