# Ansible de la cuenta de laboratorio

El workflow `aprovisionar-cuenta` genera `inventory/hosts.yml` (no se versiona) con el id de instancia y corre `site.yml`. La conexión es `amazon.aws.aws_ssm`: el puerto 22 sigue cerrado y Terraform no crea una llave SSH, porque la llave privada acabaría en el estado o se perdería con el runner efímero.

El plugin SSM necesita un cubo S3 para copiar los módulos. Ese cubo es `campus-verde-ssm-<cuenta>`, sin versionado y con expiración al día. El token del runner no viaja en el playbook: el workflow lo deja en SSM Parameter Store (`/campus/runner-token/<ambiente>`) y `scripts/runner-instalar.sh` lo lee en la instancia. Al terminar, el parámetro se borra.

Comprobar sintaxis:

```bash
ansible-playbook --syntax-check -i inventory/hosts.yml.example site.yml
```

Idempotencia del layout, en un contenedor local:

```bash
bash infra/ansible/probar-idempotencia.sh
```

La guía para quien cambia de cuenta está en [`docs/CAMBIO-DE-CUENTA-LAB.md`](../../docs/CAMBIO-DE-CUENTA-LAB.md).
