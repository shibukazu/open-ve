terraform {
  backend "s3" {
    bucket = "open-ve-terraform-state"
    key    = "key"
    region = "ap-northeast-1"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.65.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "2.4.0"
    }
  }
}

provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Terraform   = "true"
    }
  }
}
