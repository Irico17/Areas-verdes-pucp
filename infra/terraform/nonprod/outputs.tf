output "instance_id" {
  description = "ID de la instancia EC2 compartida para develop y qa"
  value       = aws_instance.nonprod.id
}

output "public_ip" {
  description = "IP publica elastica de la instancia nonprod"
  value       = aws_eip.nonprod.public_ip
}

output "volume_id" {
  description = "ID del volumen EBS persistente montado en /opt/campus"
  value       = aws_ebs_volume.data.id
}

output "urls" {
  description = "URLs publicas de los ambientes alojados en la instancia"
  value = {
    develop = "http://${aws_eip.nonprod.public_ip}:8088"
    qa      = "http://${aws_eip.nonprod.public_ip}:8188"
  }
}
