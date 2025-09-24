output "tunnel_token" {
  description = "Token for the client to authorize to the tunnel"
  value       = cloudflare_zero_trust_tunnel_cloudflared.main.tunnel_token
  sensitive   = true
}
