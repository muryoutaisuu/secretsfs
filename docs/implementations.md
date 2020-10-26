# Implementations

_secretsfs_ knows two interfaces:

* Store: Where to get the secrets from
* File Input/Output (FIOs): How to display the secrets

# Store

Store implementations are currently available for:

* Vault

# File Input/Output (FIOs)

FIO Implementations are currently the following:

|    FIO name   |                                                             Purpose                                                             |  default |
|:-------------:|:-------------------------------------------------------------------------------------------------------------------------------:|:--------:|
| secretsfiles  | To display secrets as is, just a file containing the secret.                                                                    | enabled  |
| templatefiles | To display secrets rendered into a template, e.g. a configuration file. See configuration on how to configure and use this FIO. | enabled  |
| internal      | To display some internal information of secretsfs, mostly used for debugging                                                    | enabled  |
| tests         | Used for debugging, emulating a simple FIO                                                                                      | disabled |
