# Known Issues

These are some known issues:

* **Substitution:** Prior to version 1.0.0 it was possible to substitute the '/' character in names and paths of secrets for the _secretsfiles_ FIO. I felt it too much of an edge case to have code dealing with it. Most users of IT technologies know the '/' character to be a rather bad choice to include in file names. Hence forward of version 1.0.0 the `secretsfiles` FIO will throw an error for such files. Therefore: **Do not use '/' characters in your paths and names of secrets in Vault!**
* In Vault, both paths `/secret/foo` and `/secret/foo/` may exist, where the former is a secret and the latter is a subpath. Filesystems know no difference between a path with and without the `/` at the end. Hence Both validate to the same path. In _secretsfs_ this results into the keys of `/secret/foo` being displayed as files next to the subdirectory `/secret/foo/`, while in reality those two are not connected in any way to each other in Vault. This may cause some confusion, therefore I advise to never create a secret with the same name as a path adjacent to each other in the same 'directory' in Vault.
