output "public_ip" {
  value = aws_eip.app.public_ip
}

output "public_url" {
  value = "http://${aws_eip.app.public_ip}"
}

output "instance_id" {
  value = aws_instance.app.id
}

output "ecr_api" {
  value = try(aws_ecr_repository.api[0].repository_url, "")
}

output "ecr_web" {
  value = try(aws_ecr_repository.web[0].repository_url, "")
}

output "evidence_bucket" {
  value = try(aws_s3_bucket.evidencias[0].id, "")
}

output "region" {
  value = var.aws_region
}
