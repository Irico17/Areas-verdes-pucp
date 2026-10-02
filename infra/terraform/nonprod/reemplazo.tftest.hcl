mock_provider "aws" {}

run "valida_configuracion_nonprod" {
  command = plan

  assert {
    condition     = aws_instance.nonprod.user_data_replace_on_change == false
    error_message = "user_data_replace_on_change debe quedar en false"
  }

  assert {
    condition = length([
      for rule in aws_security_group.nonprod.ingress : rule
      if rule.from_port == 22
    ]) == 0
    error_message = "sin ingress 22 por defecto"
  }

  assert {
    condition = toset([
      for rule in aws_security_group.nonprod.ingress : rule.from_port
    ]) == toset([8088, 8188])
    error_message = "SG solo con los puertos 8088/8188"
  }
}

run "valida_ssh_cidr_opcional" {
  command = plan

  variables {
    ssh_cidr = "190.119.1.5/32"
  }

  assert {
    condition = contains([
      for rule in aws_security_group.nonprod.ingress : rule.from_port
    ], 22)
    error_message = "debe permitir puerto 22 si se provee ssh_cidr"
  }
}
