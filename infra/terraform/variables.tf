variable "aws_region" {
  description = "Solo us-east-1 o us-west-2 en AWS Academy Learner Lab. us-east-1 es el default; sa-east-1 queda más cerca de Lima y no está habilitado en el lab."
  type        = string
  default     = "us-east-1"

  validation {
    condition     = contains(["us-east-1", "us-west-2"], var.aws_region)
    error_message = "Learner Lab solo permite us-east-1 y us-west-2."
  }
}

variable "instance_type" {
  description = "t3.micro cabe holgado en el presupuesto de 50 USD. t3.small si el build en la instancia se queda corto."
  type        = string
  default     = "t3.micro"

  validation {
    condition     = contains(["t3.micro", "t3.small"], var.instance_type)
    error_message = "Use t3.micro o t3.small."
  }
}

variable "instance_profile" {
  description = "Perfil que ya existe en el lab. No se crea ninguno."
  type        = string
  default     = "LabInstanceProfile"
}

variable "lab_role_name" {
  description = "Rol que ya existe. El módulo no lo crea; queda documentado para la variante ECS."
  type        = string
  default     = "LabRole"
}

variable "ssh_cidr" {
  description = "SSH solo si indica una IP/32. Vacío cierra el puerto 22. 0.0.0.0/0 no se acepta."
  type        = string
  default     = ""

  validation {
    condition     = var.ssh_cidr == "" || (can(cidrhost(var.ssh_cidr, 0)) && var.ssh_cidr != "0.0.0.0/0")
    error_message = "Deje ssh_cidr vacío o use una red que no sea 0.0.0.0/0."
  }
}

variable "ssh_key_name" {
  description = "Key pair ya existente en el lab. Vacío si solo va a usar SSM."
  type        = string
  default     = ""
}

variable "db_password" {
  description = "Clave de Postgres. Sensible, sin default. No entra al user data ni al estado: el despliegue la escribe en /opt/campus/secrets.env."
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.db_password) >= 16 && !contains(["campus-lab", "pando-local"], var.db_password)
    error_message = "TF_VAR_db_password debe tener al menos 16 caracteres y no puede ser una clave de laboratorio."
  }
}

variable "dev_password" {
  description = "Clave de las cuentas locales (CAMPUS_DEV_PASSWORD). Sensible, sin default. No entra al user data."
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.dev_password) >= 16 && !contains(["campus-lab", "pando-local"], var.dev_password)
    error_message = "TF_VAR_dev_password debe tener al menos 16 caracteres y no puede ser una clave de laboratorio."
  }
}

variable "create_ecr" {
  description = "Crea dos repositorios ECR. El lab suele permitirlo; no crea roles."
  type        = bool
  default     = true
}

variable "api_image" {
  description = "Imagen de la API. Vacío usa el repositorio ECR de este stack, tag latest."
  type        = string
  default     = ""
}

variable "web_image" {
  description = "Imagen de nginx. Vacío usa el repositorio ECR de este stack, tag latest."
  type        = string
  default     = ""
}

variable "create_evidence_bucket" {
  description = "Cubo S3 opcional para evidencias. La instancia lo usa con LabRole. Apagado por defecto."
  type        = bool
  default     = false
}

variable "enable_ecs" {
  description = "No aplica la variante ECS en este root. Quédese en false en Learner Lab. Ver modules/ecs-fargate."
  type        = bool
  default     = false
}
