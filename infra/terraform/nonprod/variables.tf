variable "aws_region" {
  description = "Región de AWS Learner Lab (us-east-1 o us-west-2)."
  type        = string
  default     = "us-east-1"

  validation {
    condition     = contains(["us-east-1", "us-west-2"], var.aws_region)
    error_message = "Learner Lab solo permite us-east-1 y us-west-2."
  }
}

variable "instance_type" {
  description = "Tipo de instancia EC2 para la máquina compartida develop/qa. t3.small es el default para cuidar el crédito; t3.medium si los dos stacks se quedan cortos de memoria."
  type        = string
  default     = "t3.small"

  validation {
    condition     = contains(["t3.micro", "t3.small", "t3.medium"], var.instance_type)
    error_message = "Use t3.micro, t3.small o t3.medium."
  }
}

variable "root_volume_gb" {
  description = "Disco raíz gp3. Bajar el tamaño de un volumen ya creado lo reemplaza."
  type        = number
  default     = 20

  validation {
    condition     = var.root_volume_gb >= 20 && var.root_volume_gb <= 40
    error_message = "root_volume_gb debe estar entre 20 y 40."
  }
}

variable "data_volume_gb" {
  description = "Tamaño en GB del volumen EBS persistente gp3 montado en /opt/campus para develop y qa. La máquina anterior usaba 40; bajarlo reemplaza el disco y borra los datos."
  type        = number
  default     = 20

  validation {
    condition     = var.data_volume_gb >= 20 && var.data_volume_gb <= 40
    error_message = "data_volume_gb debe estar entre 20 y 40."
  }
}

variable "instance_profile" {
  description = "Perfil de instancia IAM preexistente en Learner Lab."
  type        = string
  default     = "LabInstanceProfile"
}

variable "ssh_cidr" {
  description = "SSH solo si indica una IP/32 o CIDR específico. Vacío cierra el puerto 22 (administración mediante SSM). 0.0.0.0/0 no se acepta."
  type        = string
  default     = ""

  validation {
    condition     = var.ssh_cidr == "" || (can(cidrhost(var.ssh_cidr, 0)) && var.ssh_cidr != "0.0.0.0/0")
    error_message = "Deje ssh_cidr vacío o use una red que no sea 0.0.0.0/0."
  }
}

variable "web_ports" {
  description = "Puertos TCP públicos para las aplicaciones web de los ambientes alojados (8088 develop, 8188 qa)."
  type        = list(number)
  default     = [8088, 8188]
}
