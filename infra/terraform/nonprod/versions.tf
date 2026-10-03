terraform {
  required_version = ">= 1.10.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
  }

  # Misma cuenta y mismo cubo que el root de producción; otra clave de estado.
  # Cubo y región llegan por -backend-config (ver scripts/bootstrap-estado-terraform.sh).
  backend "s3" {
    key          = "learner-lab-nonprod/terraform.tfstate"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project       = "campus-verde"
      Stack         = "nonprod"
      CampusGestion = "cuenta-lab"
    }
  }
}
