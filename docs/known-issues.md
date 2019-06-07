# Known Issues

These are some known issues:

* **Substitution:** `a/b` may be substituted to `a_b`, which may also already be in the backend (e.g. Vault). This will likely cause a clash. As a workaround either configure `subst_char` in the configuration file to a different value, or do not use `/` in Vault key names at all. If clashing, the alphabetically first key name will have precedence (in this case it would be `a/b`).
* In Vault, both paths `/secret/foo` and `/secret/foo/` may exist, where the former is a secret and the latter is a subpath. Filesystems know no difference between a path with and one without the `/` at the end. Hence Both validate to the same path. In _secretsfs_ this results into the keys of `/secret/foo` being displayed as files next to the subdirectory `/secret/foo/`, while in reality those two are not connected in any way to each other in Vault.