import * as path from "path";
import * as fs from "fs";

import * as nomad from "@pulumi/nomad";
import * as vault from "@pulumi/vault";

const _jobs: Record<string, { path: string }> = {
  "acme-example": {
    path: path.join(__dirname, "job.hcl"),
  },
};

Object.entries(_jobs).reduce((acc, [key, attributes]) => {
  let volume = undefined;

  acc[key] = new nomad.Job(
    key,
    {
      jobspec: fs.readFileSync(attributes.path).toString(),
      detach: false,
      hcl2: {
        enabled: true,
        allowFs: true,
      },
    },
    { dependsOn: volume }
  );
  return acc;
}, {} as Record<string, nomad.Job>);

new vault.Policy("acme-example", {
  name: "acme-example",
  policy: fs.readFileSync(path.join(__dirname, "policy.hcl")).toString(),
});

new vault.generic.Endpoint("acme-acme-example-role", {
  // TODO: `acme` should be a variable
  path: `acme/roles/acme-example`,
  disableRead: true,
  disableDelete: false,
  dataJson: JSON.stringify({
    account: "letsencrypt-staging",
    allowed_domains: "acme-example.prod.stratos.host",
    allow_subdomains: true,
    allow_bare_domains: false,
  }),
});
