#!/bin/bash

if [ "$1" = "-h" ]; then
 cat 1>&2 <<EOL
usage: echo /etc/passwd | ./paste.sh

env options:
  PASTE_URL  - Set the url base for paste operations (default: HTTPS://paste.dn42.us)
  PASTE_GZIP - 0 = No Compression,  1 = Use gzip compression (default: 0)
  PASTE_BURN - 0 = No Burn on Read, 1 = Burn on read         (default: 0)
  PASTE_DATE - Value to be used when setting expire date.    (default: next-week)
EOL
  exit
fi

PASTE_URL=${PASTE_URL-"https://paste.dn42.us"}
PASTE_BURN=${PASTE_BURN-0}
PASTE_DATE=${PASTE_DATE-"next-year"}
PASTE_GZIP=${PASTE_GZIP-0}
GZBIN="cat"
[ "$PASTE_GZIP" -eq "1" ] && GZBIN="gzip -c"
    TS="$(date +%s -d ${PASTE_DATE})"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    TS="$(date +%s -d ${PASTE_DATE})"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ "$PASTE_DATE" = "next-year" ]]; then
        PASTE_DATE="+1y"
    elif [[ "$PASTE_DATE" = "next-month" ]]; then
        PASTE_DATE="+1m"
    elif [[ "$PASTE_DATE" = "next-week" ]]; then
        PASTE_DATE="+1m"
    fi

    TS="$(date +%s -v ${PASTE_DATE})"
fi

PASS=$(head  -c 40 /dev/urandom);
CHK=$(echo   -s $PASS | openssl dgst -sha256 -binary | openssl dgst -ripemd160 -binary | base64 | tr '/+' '_-' | tr -d '=')
PASS=$(echo  -s $PASS | openssl dgst -sha256 -binary | base64 | tr '/+' '_-' | tr -d '=')
HASH=$((echo -e "exp:\t$TS"; \
        echo -e "chk:\t$CHK"; \
        [ "$PASTE_BURN" -eq "1" ] && echo -e "burn:\ttrue"; \
        [ "$PASTE_GZIP" -eq "1" ] && echo -e "zip:\ttrue"; \
        echo; \
        cat /dev/stdin | $GZBIN | openssl aes-256-cbc -md md5 -e -a -k $PASS) | \
        curl -s -X POST ${PASTE_URL}/paste --data-binary @-)

HASH_OK=$(echo $HASH | cut -c1-2)

if [ "$HASH_OK" = "OK" ]; then
  HASH=$(echo $HASH | cut -f2 -d' ')

  echo "url: ${PASTE_URL}/#/${HASH}!${PASS}"
  echo -n "shell: curl -s ${PASTE_URL}/api/get/${HASH} | sed '1,/^\$/d' | openssl aes-256-cbc -md md5 -d -a -k ${PASS}"
  [ "$PASTE_GZIP" -eq "1" ] && echo " | gzip -dc" || echo;
  exit
fi

echo $HASH