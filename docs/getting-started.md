# Just Test _secretsfs_

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

# Install _secretsfs_

## From Source

To get started with _secretsfs_ simply download this repository and install it:

```bash
go get github.com/muryoutaisuu/secretsfs            # get main source
cd $GOPATH/src/github.com/muryoutaisuu/secretsfs    # change to directory
go get ./...                                        # get dependencies
go install ./cmd/secretsfs                          # install secretsfs
```

## Just Copy the Binary

For Linux x86_64, you may download the latest built binary from [the project's release page](https://github.com/muryoutaisuu/secretsfs/releases).

## RPM Package

For RHEL7, there is a prebuilt .rpm package, you may download from [the project's release page](https://github.com/muryoutaisuu/secretsfs/releases).

# Start *secretsfs*

There are two possible ways to start *secretsfs*:

Either start it manually with `secretsfs <mountpath> [-o <mountoptions>] [&]]`, or start it with Systemd using the predefined service in the examples folder:

```bash
cp example/secretsfs.service /usr/lib/systemd/system/secretsfs.service
systemctl start secretsfs
systemctl enable secretsfs
```

The Systemd definition also comes with your .rpm package installation.
