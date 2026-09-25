data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_caller_identity" "current" {}

resource "aws_ecr_repository" "api" {
  count                = var.create_ecr ? 1 : 0
  name                 = "campus-verde-api"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_repository" "web" {
  count                = var.create_ecr ? 1 : 0
  name                 = "campus-verde-web"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_security_group" "app" {
  name        = "campus-verde-app"
  description = "HTTP, HTTPS reservado y SSH del piloto Campus Verde"

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS reservado"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "ingress" {
    for_each = var.ssh_cidr == "" ? [] : [var.ssh_cidr]
    content {
      description = "SSH restringido a una red"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "app" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = var.instance_type
  iam_instance_profile   = var.instance_profile
  vpc_security_group_ids = [aws_security_group.app.id]
  key_name               = var.ssh_key_name != "" ? var.ssh_key_name : null

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
  }

  user_data = templatefile("${path.module}/user_data.sh.tftpl", {
    api_image       = local.api_image
    web_image       = local.web_image
    registry        = local.registry
    region          = var.aws_region
    evidence_bucket = var.create_evidence_bucket ? aws_s3_bucket.evidencias[0].id : ""
  })

  # Un cambio de user_data (por ejemplo el tag de la imagen) se aplica in-place.
  # cloud-init no se repite: el despliegue es docker compose pull, no un reemplazo.
  user_data_replace_on_change = false

  lifecycle {
    ignore_changes = [ami]
  }

  metadata_options {
    http_endpoint = "enabled"
    http_tokens   = "required"
  }

  tags = {
    Name = "campus-verde"
  }
}

resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.app.availability_zone
  size              = 20
  type              = "gp3"

  tags = {
    Name = "campus-verde-data"
  }
}

resource "aws_volume_attachment" "data" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.app.id
}

resource "aws_eip" "app" {
  domain = "vpc"

  tags = {
    Name = "campus-verde"
  }
}

resource "aws_eip_association" "app" {
  instance_id   = aws_instance.app.id
  allocation_id = aws_eip.app.id
}

resource "aws_s3_bucket" "evidencias" {
  count  = var.create_evidence_bucket ? 1 : 0
  bucket = "campus-verde-ev-${data.aws_caller_identity.current.account_id}"
}

resource "aws_s3_bucket_versioning" "evidencias" {
  count  = var.create_evidence_bucket ? 1 : 0
  bucket = aws_s3_bucket.evidencias[0].id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "evidencias" {
  count                   = var.create_evidence_bucket ? 1 : 0
  bucket                  = aws_s3_bucket.evidencias[0].id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

locals {
  api_image = var.api_image != "" ? var.api_image : (var.create_ecr ? "${aws_ecr_repository.api[0].repository_url}:latest" : "")
  web_image = var.web_image != "" ? var.web_image : (var.create_ecr ? "${aws_ecr_repository.web[0].repository_url}:latest" : "")
  registry  = try(split("/", local.api_image)[0], "")
}

check "ecs_no_es_el_default" {
  assert {
    condition     = var.enable_ecs == false
    error_message = "enable_ecs debe seguir en false. La variante ECS está en modules/ecs-fargate y no se aplica desde este root."
  }
}
