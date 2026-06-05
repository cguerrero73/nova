# PgBouncer Setup

PgBouncer corre junto a PostgreSQL para multiplexar conexiones.

## Topología

```
Go ──► :6432 ──► PgBouncer ──► :5432 ──► PostgreSQL
           (este server)        localhost
```

## Instalación en el servidor de DB

### Opción 1: Paquete del sistema (recomendado)

```bash
# Debian/Ubuntu
sudo apt install pgbouncer

# RHEL/CentOS
sudo yum install pgbouncer
```

### Opción 2: Docker

```bash
docker run -d \
  --name pgbouncer \
  -p 6432:5432 \
  -e DATABASE_URL="postgres://user:pass@127.0.0.1:5432/nova" \
  -e POOL_MODE=session \
  -e MAX_CLIENT_CONN=100 \
  -e DEFAULT_POOL_SIZE=20 \
  edoburu/pgbouncer:latest
```

## Configuración

1. Copiar archivos de configuración:

```bash
sudo cp pgbouncer.ini /etc/pgbouncer/pgbouncer.ini
sudo cp userlist.txt /etc/pgbouncer/userlist.txt
```

2. Editar `pgbouncer.ini` con las credenciales reales de PostgreSQL:

```ini
[databases]
nova = host=127.0.0.1 port=5432 dbname=nova user=TU_USER password=TU_PASS
```

3. Generar hashes para userlist.txt:

```bash
# Para cada usuario
./generate_userlist.sh dev dev
```

4. Dar permisos:

```bash
sudo chown pgbouncer:pgbouncer /etc/pgbouncer/*.txt
sudo chmod 640 /etc/pgbouncer/*.txt
```

## Iniciar

```bash
# Sistema (systemd)
sudo systemctl enable pgbouncer
sudo systemctl start pgbouncer
sudo systemctl status pgbouncer

# O manualmente
pgbouncer -d /etc/pgbouncer/pgbouncer.ini
```

## Verificar

```bash
# Ver que corre
ps aux | grep pgbouncer

# Test de conexión
psql -h 127.0.0.1 -p 6432 -U dev -d nova

# Test desde otra máquina
psql -h <IP_SERVIDOR> -p 6432 -U dev -d nova
```

## Logs

```bash
# Ver logs
sudo tail -f /var/log/pgbouncer/pgbouncer.log

# Stats
echo "SHOW CLIENTS;" | psql -h 127.0.0.1 -p 6432 -U dev pgbouncer
echo "SHOW POOLS;" | psql -h 127.0.0.1 -p 6432 -U dev pgbouncer
```

## Parámetros importantes

| Parámetro           | Valor     | Descripción                                                   |
| ------------------- | --------- | ------------------------------------------------------------- |
| `pool_mode`         | `session` | Mantiene search_path por sesión (requerido para multi-tenant) |
| `max_client_conn`   | `100`     | Máx conexiones desde clientes                                 |
| `default_pool_size` | `20`      | Conexiones reales a PostgreSQL                                |
| `listen_port`       | `6432`    | Puerto donde escucha PgBouncer                                |

## Troubleshooting

### "login failed"

- Verificar que userlist.txt tenga los hashes correctos
- Regenerar hash: `echo -n "passwordusername" | md5sum`

### "cannot connect to server"

- Verificar PostgreSQL corriendo: `pg_isready -h 127.0.0.1 -p 5432`
- Ver logs de PgBouncer
- Firewall permite conexiones al puerto 6432
