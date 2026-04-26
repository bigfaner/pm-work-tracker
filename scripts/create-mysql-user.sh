#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 <username> [--host localhost] [--database dbname]"
  echo "  Creates a MySQL user with a random password and grants full access to a database."
  echo ""
  echo "Options:"
  echo "  --host HOST       MySQL host (default: localhost)"
  echo "  --database DB     Database to grant access to (default: same as username)"
  echo "  --root-user USER  Root user for MySQL (default: root)"
  echo "  --root-pass PASS  Root password (prompted if not provided)"
  exit 1
}

USERNAME=""
HOST="localhost"
DATABASE=""
ROOT_USER="root"
ROOT_PASS=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --host) HOST="$2"; shift 2 ;;
    --database) DATABASE="$2"; shift 2 ;;
    --root-user) ROOT_USER="$2"; shift 2 ;;
    --root-pass) ROOT_PASS="$2"; shift 2 ;;
    -h|--help) usage ;;
    -*)
      echo "Unknown option: $1"
      usage
      ;;
    *)
      if [[ -z "$USERNAME" ]]; then
        USERNAME="$1"
      else
        echo "Unexpected argument: $1"
        usage
      fi
      shift
      ;;
  esac
done

if [[ -z "$USERNAME" ]]; then
  echo "Error: username is required"
  usage
fi

if [[ -z "$DATABASE" ]]; then
  DATABASE="${USERNAME}_db"
fi

# Generate random 16-char password (alphanumeric)
PASSWORD=$(openssl rand -hex 8)

# Build mysql command
MYSQL_ARGS=(-u "$ROOT_USER" -h "$HOST")
if [[ -n "$ROOT_PASS" ]]; then
  MYSQL_ARGS+=(-p"$ROOT_PASS")
fi

echo "Creating MySQL user: ${USERNAME}@${HOST}"
echo "Database: ${DATABASE}"
echo "Generated password: ${PASSWORD}"
echo ""

mysql "${MYSQL_ARGS[@]}" -e "
  CREATE USER IF NOT EXISTS '${USERNAME}'@'${HOST}' IDENTIFIED BY '${PASSWORD}';
  CREATE DATABASE IF NOT EXISTS \`${DATABASE}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
  GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP ON \`${DATABASE}\`.* TO '${USERNAME}'@'${HOST}';
  FLUSH PRIVILEGES;
"

echo "Done. Connection string:"
echo "  ${USERNAME}:${PASSWORD}@tcp(${HOST}:3306)/${DATABASE}"
