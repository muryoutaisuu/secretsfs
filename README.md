[![Build Status](https://travis-ci.com/muryoutaisuu/secretsfs.svg?branch=master)](https://travis-ci.com/muryoutaisuu/secretsfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/muryoutaisuu/secretsfs)](https://goreportcard.com/report/github.com/muryoutaisuu/secretsfs)
[![Documentation Status](https://readthedocs.org/projects/secretsfs/badge/?version=latest)](https://secretsfs.readthedocs.io/en/latest/?badge=latest)
[![GoDoc](https://godoc.org/github.com/muryoutaisuu/secretsfs?status.svg)](https://godoc.org/github.com/muryoutaisuu/secretsfs/pkg)

* [GoDoc](https://godoc.org/github.com/muryoutaisuu/secretsfs/pkg)
* [GoWalker](https://gowalker.org/github.com/muryoutaisuu/secretsfs/pkg)


# Work in Progress...

Subject to changes of ownership, api and code structure amongst other things...

**Do not use yet!!!**

# Purpose of *secretsfs*

_secretsfs_ implements a FUSE-filesytem, that allows to interact with secrets stored in a backend (called store) via simple filesysten-interacting commands.
One such store may be [Vault](https://github.com/hashicorp/vault).

Output formats (called FIO, stands for File Input/Output) are treated like plugins and can be (de-)activated in a configuration file. Out of the box implemented FIOs are:

* **secretsfiles:** returns plain secret on a simple `cat`
* **templatefiles:** returns on `cat` a with secrets rendered file (e.g. a configuration file with secrets)

[Read the docs for more!](https://secretsfs.readthedocs.io/en/latest/)
