locals {
  loadchecker_image = "ghcr.io/bloominlabs/loadchecker:latest"
}

job "loadchecker" {
  datacenters = ["do-sfo3"]

  group "loadchecker" {
    network {
      dns {
        servers = ["172.17.0.1", "1.1.1.1", "1.0.0.1"]
      }

      port "http" {
        host_network = "tailscale"
      }
    }

    service {
      name = "loadchecker"
      port = "http"
      tags = ["http"]

      check {
        name            = "http"
        type            = "http"
        protocol        = "http"
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

    task "loadchecker" {
      driver = "docker"

      config {
        image = local.loadchecker_image
        ports = ["http"]
        args = [
          "-otlp.addr", "172.17.0.1:4317",
          "-otlp.cert.path", "${NOMAD_SECRETS_DIR}/observability.cert.pem",
          "-otlp.key.path", "${NOMAD_SECRETS_DIR}/observability.key.pem",
          "-otlp.ca.path", "${NOMAD_SECRETS_DIR}/observability.ca.pem"
        ]
      }

      resources {
        cpu    = 128
        memory = 32
      }

      vault {
        policies = ["loadchecker"]
      }

      // certificate for authenticating to baremetal-otel-collector
      template {
        data        = <<EOH
{{- $ip_sans := printf "ip_sans=%s" (env "attr.unique.network.ip-address") -}}
{{- with secret "observability_intermediate/issue/loadchecker" "common_name=loadchecker.service.consul" $ip_sans -}}
{{- range $index, $value := .Data.ca_chain -}}
{{- printf "%s\n" $value -}}
{{- end -}}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/observability.ca.pem"
        perms       = "700"
      }

      template {
        data        = <<EOH
{{ $ip_sans := printf "ip_sans=%s" (env "attr.unique.network.ip-address") -}}
{{ with secret "observability_intermediate/issue/loadchecker" "common_name=loadchecker.service.consul" $ip_sans -}}
{{ .Data.certificate }}
{{- range $index, $value := .Data.ca_chain }}
{{ $value }}
{{- end }}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/observability.cert.pem"
        perms       = "700"
      }

      template {
        data        = <<EOH
{{- $ip_sans := printf "ip_sans=%s" (env "attr.unique.network.ip-address") }}
{{- with secret "observability_intermediate/issue/loadchecker" "common_name=loadchecker.service.consul" $ip_sans -}}
{{- .Data.private_key -}}
{{- end -}}
EOH
        destination = "${NOMAD_SECRETS_DIR}/observability.key.pem"
        perms       = "700"
      }
    }
  }
}
