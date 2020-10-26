# secretsfs

**TL;DR**: Access your secrets securely via a simple `cat` command instead of using a client.

_secretsfs_ implements a [FUSE](https://en.wikipedia.org/wiki/Filesystem_in_Userspace)-filesytem, that allows you to interact with secrets stored in a backend (called store) via simple readonly filesysten-interacting commands, like cat, grep etc.
One such store may be [Vault](https://github.com/hashicorp/vault).

Output formats _File Input/Output_ (FIO) are treated like plugins and can be (de-)activated in a configuration file. Out of the box implemented FIOs are:

* **secretsfiles:** returns plain secret on a simple `cat`
* **templatefiles:** returns on `cat` a with secrets rendered file (e.g. a configuration file with secrets)
* **internal:** mostly used for checking the state of _secretsfs_ and debugging
* **tests:** disabled by default, mostly used for unit testing

Get it now on [GitHub](https://github.com/muryoutaisuu/secretsfs)!
