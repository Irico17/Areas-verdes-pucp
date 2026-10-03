terraform {
  required_version = ">= 1.10.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
  }

  # El cubo y la región no van en el código: cada cuenta de Learner Lab
  # tiene el suyo (campus-verde-tfstate-<id de la cuenta>). Lo crea
  # scripts/bootstrap-estado-terraform.sh y se pasa con -backend-config.
  # use_lockfile evita una tabla DynamoDB, que este lab no necesita.
  backend "s3" {
    key          = "learner-lab/terraform.tfstate"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project       = "campus-verde"
      Stack         = "learner-lab"
      Ambiente      = var.ambiente
      CampusGestion = "cuenta-lab"
    }
  }
}
