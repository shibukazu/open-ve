
module "load_test" {
  source = "./modules/aws/load_test"
}

module "oidc" {
  source = "./modules/aws/oidc"
}

output "load_test_ecr_repository_url" {
  value = module.load_test.ecr_repository_url
}

output "load_test_public_subnet" {
  value = module.load_test.public_subnet
}

output "load_test_vpc_id" {
  value = module.load_test.vpc_id
}

output "load_test_ecs_cluster_name" {
  value = module.load_test.ecs_cluster_name
}

output "load_test_ecs_task_definition" {
  value = module.load_test.ecs_task_definition
}

output "load_test_service_security_group_id" {
  value = module.load_test.service_security_group_id
}
