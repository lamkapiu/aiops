module "cvm" {
  source     = "../module/cvm"
  secret_id  = var.secret_id
  secret_key = var.secret_key
  password   = var.password
}
