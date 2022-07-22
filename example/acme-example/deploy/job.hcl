locals {
  acme-example_image = "ghcr.io/bloominlabs/acme-example:latest"
}

job "acme-example" {
  datacenters = ["do-sfo3"]
  type        = "service"

  group "acme-example" {
    network {
      dns {
        servers = ["172.17.0.1", "1.1.1.1", "1.0.0.1"]
      }

      port "http" {
        host_network = "public"
      }
    }

    service {
      name = "${JOB}"
      port = "http"
      tags = ["public"]

      check {
        name     = "http"
        type     = "http"
        protocol = "https"
        // This is needed in dev but not in prod because we're using the
        // letsencrypt staging certificate which isn't trusted publicly. We can fix
        // this in the test environment when they're completely seperate
        tls_skip_verify = true
        path            = "/"
        interval        = "5s"
        timeout         = "2s"

        check_restart {
          limit           = 3
          grace           = "90s"
          ignore_warnings = false
        }
      }
    }

    task "acme-example" {
      driver = "docker"

      vault {
        policies = ["acme-example"]
      }

      config {
        image = local.acme-example_image
        ports = ["http"]
      }

      resources {
        cpu = 10
        # TODO: figure out how to reduce memory. on disk buffer?
        memory = 10
      }

      template {
        data = <<EOH
    TLS_CERT_PATH="{{ env "NOMAD_SECRETS_DIR" }}/server.cert.pem"
    TLS_KEY_PATH="{{ env "NOMAD_SECRETS_DIR" }}/server.key.pem"
    EOH

        destination = "${NOMAD_SECRETS_DIR}/.env"
        env         = true
      }

      // TODO: the common_name should likely not be hard coded since we dont
      // always want it to be prod in the future, but going to leave it for now
      template {
        data        = <<EOH
{{- with secret "acme/certs/acme-example" "common_name=acme-example.prod.stratos.host" -}}
{{- .Data.issuer_cert -}}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/server.ca.pem"
        perms       = "700"

        change_mode   = "signal"
        change_signal = "SIGINT"
      }

      template {
        data        = <<EOH
{{- with secret "acme/certs/acme-example" "common_name=acme-example.prod.stratos.host" -}}
{{- .Data.cert -}}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/server.cert.pem"
        perms       = "700"

        change_mode   = "signal"
        change_signal = "SIGINT"
      }

      template {
        data        = <<EOH
{{- with secret "acme/certs/acme-example" "common_name=acme-example.prod.stratos.host" -}}
{{- .Data.private_key -}}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/server.key.pem"
        perms       = "700"

        change_mode   = "signal"
        change_signal = "SIGINT"
      }
    }
  }
}
