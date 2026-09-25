# Un apply que solo cambia la imagen no reemplaza la EC2.
# Se ejecuta con el proveedor simulado: no llama a AWS.

mock_provider "aws" {}

variables {
  db_password  = "clave-postgres-prueba-16"
  dev_password = "clave-cuentas-prueba-16"
  api_image    = "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-api:uno"
  web_image    = "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-web:uno"
}

run "crea" {
  command = apply

  assert {
    condition     = aws_instance.app.user_data_replace_on_change == false
    error_message = "user_data_replace_on_change debe quedar en false"
  }
}

run "cambia_imagen" {
  command = apply

  variables {
    api_image = "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-api:dos"
  }

  assert {
    condition     = aws_instance.app.id == run.crea.instance_id
    error_message = "cambiar solo la imagen reemplazó la instancia"
  }

  assert {
    condition     = aws_instance.app.user_data_replace_on_change == false
    error_message = "el segundo apply volvió a pedir reemplazo por user data"
  }
}
