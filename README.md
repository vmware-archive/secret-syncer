# concourse secret syncer

This is a little tool for using a structured YAML file as the source of truth
for the set of credentials that should exist in a Vault instance to be
accessed by concourse.

You declare your secrets like so:
```yaml
# pipeline-scoped secrets
team_name/pipeline_name/secret_name: simple_value
team_name/pipeline_name/other_secret_name:
  compound: multi
  field: value
# shared secrets
shared:
  secret_name: secret_value
  other_secret:
    compound: value
    for: other_secret
# team-scoped secrets
team_name/secret_name: value
```

you configure the program using the same environment variables that would be
used to target a Vault instance with the `vault` CLI:

* `VAULT_ADDR`
* `VAULT_TOKEN`
* ... etc

The program will recursively delete all the secrets under the path `concourse/*`
and then write every secret declared in the YAML file (in the above format)
passed as a positional argument:

```
VAULT_ADDR=http://127.0.0.1 VAULT_TOKEN=myroot go run main.go secrets.yml
```
