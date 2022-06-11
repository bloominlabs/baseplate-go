# acme-example

Example project for using the [`acme/` vault
mount](https://github.com/remilapeyre/vault-acme/) for generating [letsencrypt
staging certificates](https://letsencrypt.org/docs/staging-environment/) and
automatically rotating them from disk receiving SIGHUP.

This example isn't the 'best practices' yet as we work to develop those so a couple of things to watch out for:

- we should add some basic baseplate metrics for the TLS certificates to catch if they expire and alert
- we haven't fully tested the `acme/` endpoint so it may have some weird
  behaviors. One thing I was seeing is that it caches the same certificate for
  every request for a particular domain name. This is good for our ratelimit,
  but base since all of the certificates will expire at the exact same time
  across the entire infrastructure and may lead to revocation problems.
- there are some todos in the various files of things to improve.
- need to add a basic test framework to make sure certificates are reloaded from disk
- we may want to consider using a [file watcher
  instead](https://github.com/hashicorp/consul/pull/12329/files) or something
  like
  <https://github.com/hashicorp/go-secure-stdlib/blob/main/reloadutil/reload.go>

The nice thing about this setup however is since we use consul-template / nomad
to write the letsencrypt certificates to disk, you can test the example code (and write tests)
using any certificate that you want (`internal` which is already trusted for
accesing vault/consul, self-signed, etc.)

Play around with the main.go and deploy to see whats going on.
