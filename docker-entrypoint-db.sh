#!/bin/bash
set -euxo pipefail

/db pretreatment -dbname $DB_NAME -dburl $POSTGRESQL_URL

exec /db up -dburl $POSTGRESQL_RUNE_OIDC_URL