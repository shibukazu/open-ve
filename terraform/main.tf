
module "load_test" {
  source        = "./modules/aws/load_test"
  preshared_key = var.preshared_key
}

module "oidc" {
  source = "./modules/aws/oidc"
}

output "load_test_ecr_repository_url" {
  value = module.load_test.ecr_repository_url
}

output "load_test_public_subnets" {
  value = module.load_test.public_subnets
}

output "load_test_vpc_id" {
  value = module.load_test.vpc_id
}
