variable "cloudflare_account_id" {
  description = "ID of the Cloudflare account"
  type        = string
}

variable "cloudflare_api_token" {
  description = "Token to connect to the Cloudflare API"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "ID of the Cloudflare hosted zone for the domain"
  type        = string
}

variable "domain" {
  type = string
}
