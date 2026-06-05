#!/bin/bash
# Genera userlist.txt con passwords hasheados para PgBouncer
# Uso: ./generate_userlist.sh user password [output_file]

USER="$1"
PASS="$2"
OUTPUT="${3:-userlist.txt}"

if [ -z "$USER" ] || [ -z "$PASS" ]; then
  echo "Uso: $0 <usuario> <password> [output_file]"
  echo "Genera el hash MD5 de la password y lo agrega al userlist.txt"
  exit 1
fi

# Generar hash MD5: md5(password + username)
HASH=$(echo -n "${PASS}${USER}" | md5sum | cut -d' ' -f1)

echo "\"$USER\" \"$HASH\"" >>"$OUTPUT"
echo "Agregado: $USER -> $HASH"
echo "Archivo: $OUTPUT"
