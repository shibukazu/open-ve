
module "load_test" {
  source        = "./modules/aws/load_test"
  preshared_key = var.preshared_key
}

module "oidc" {
  source = "./modules/aws/oidc"
}
