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

resource "aws_security_group" "nonprod" {
  name        = "campus-verde-nonprod"
  description = "Web develop (8088), qa (8188) y SSH opcional para nonprod"

  dynamic "ingress" {
    for_each = var.web_ports
    content {
      description = "Web port ${ingress.value}"
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.ssh_cidr == "" ? [] : [var.ssh_cidr]
    content {
      description = "SSH restringido a una red especifica"
      from_port   = 22
      to_port     = 22
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    description = "Trafico de salida total"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "campus-verde-nonprod"
  }
}

resource "aws_instance" "nonprod" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = var.instance_type
  iam_instance_profile   = var.instance_profile
  vpc_security_group_ids = [aws_security_group.nonprod.id]

  root_block_device {
    volume_type = "gp3"
    volume_size = 30
    encrypted   = true
  }

  user_data = templatefile("${path.module}/bootstrap.sh.tftpl", {})

  user_data_replace_on_change = false

  lifecycle {
    ignore_changes = [ami, user_data]
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  tags = {
    Name = "campus-verde-nonprod"
  }
}

resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.nonprod.availability_zone
  size              = var.data_volume_gb
  type              = "gp3"
  encrypted         = true

  tags = {
    Name = "campus-verde-nonprod-data"
  }
}

resource "aws_volume_attachment" "data" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.nonprod.id
}

resource "aws_eip" "nonprod" {
  domain = "vpc"

  tags = {
    Name = "campus-verde-nonprod"
  }
}

resource "aws_eip_association" "nonprod" {
  instance_id   = aws_instance.nonprod.id
  allocation_id = aws_eip.nonprod.id
}
