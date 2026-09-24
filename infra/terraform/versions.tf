terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
  }

  # Estado en S3 para que una laptop con credenciales nuevas del lab pueda
  # actualizar o destruir. El cubo se crea a mano (versionado, SSE-S3);
  # este root no crea IAM ni una tabla de lock.
  backend "s3" {
    bucket  = "campus-verde-tfstate-890991908027"
    key     = "learner-lab/terraform.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project = "campus-verde"
      Stack   = "learner-lab"
    }
  }
}
