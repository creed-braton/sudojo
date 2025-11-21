resource "random_password" "tunnel_secret" {
  length = 64
}

resource "cloudflare_zero_trust_tunnel_cloudflared" "main" {
  account_id = var.cloudflare_account_id
  name       = "sudojo-tunnel"
  secret     = base64sha256(random_password.tunnel_secret.result)
}

resource "cloudflare_record" "main" {
  zone_id = var.cloudflare_zone_id
  name    = "@"
  content = cloudflare_zero_trust_tunnel_cloudflared.main.cname
  type    = "CNAME"
  proxied = true
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "main" {
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.main.id
  account_id = var.cloudflare_account_id
  config {
    ingress_rule {
      hostname = cloudflare_record.main.hostname
      service  = "http://proxy-service.sudojo.svc.cluster.local:80"
      origin_request {
        connect_timeout = "3m0s"
      }
    }
    ingress_rule {
      service = "http_status:404"
    }
  }
}
