#!/bin/bash
set -eo pipefail

REPORT=$1

PCT=$(grep total "$REPORT"| awk '{print $3}' | cut -d '.' -f1)

if [ -z "$SECRET" ]
then
	echo Missing Secret >&2
	exit 1
fi

if [ -z "$ID" ]
then
    ID="$DRONE_REPO_NAMESPACE.$DRONE_REPO_NAME"
fi

color=inactive
pct="n/a"
if ((PCT >= 0 && PCT <= 40)); then
	color=red
	pct="$PCT%25"
elif ((PCT <= 50)); then
	color=orange
	pct="$PCT%25"
elif ((PCT <= 60)); then
	color=yellow
	pct="$PCT%25"
elif ((PCT <= 70)); then
	color=yellowgreen
	pct="$PCT%25"
elif ((PCT <= 80)); then
	color=green
	pct="$PCT%25"
elif ((PCT <= 90)); then
	color=brightgreen
	pct="$PCT%25"
fi

http --ignore-stdin --form PUT "https://paste.dn42.us/s/cover.${ID}.badge" \
  url="https://img.shields.io/badge/Coverage-${pct}-${color}?style=for-the-badge&logo=go" \
  secret="$SECRET"

echo "[![Coverage Report](https://paste.dn42.us/s/cover.${ID}.badge)](https://paste.dn42.us/s/cover.${ID}.report)" > report.txt
echo '```txt' >> report.txt
cat "$REPORT" >> report.txt
echo '```' >> report.txt
PASTE_LANG="markdown"
PASTE_FILE=report.txt


PASTE_FILE=${PASTE_FILE-/dev/stdin}
PASTE_URL=${PASTE_URL-"https://paste.dn42.us"}
PASTE_BURN=${PASTE_BURN-0}
PASTE_DATE=${PASTE_DATE-""}
PASTE_GZIP=${PASTE_GZIP-1}
PASTE_LANG=${PASTE_LANG-text}
GZBIN="cat"
[ "$PASTE_GZIP" -eq "1" ] && GZBIN="gzip -c"

OS=$(uname)
case $OS in
  'Linux')
    OS='Linux'
    TS="$(date +%s -d ${PASTE_DATE})"
    ;;
  'Darwin')
    OS='Mac'
    if [[ "$PASTE_DATE" = "next-year" ]]; then
        PASTE_DATE="+1y"
    elif [[ "$PASTE_DATE" = "next-month" ]]; then
        PASTE_DATE="+1m"
    elif [[ "$PASTE_DATE" = "next-week" ]]; then
        PASTE_DATE="+1m"
    fi

    TS="$(date -v ${PASTE_DATE} +%s)"
    ;;
esac

PASS=$(head  -c 40 /dev/urandom);
CHK=$(echo   -s $PASS | openssl dgst -sha256 -binary | openssl dgst -ripemd160 -binary | base64 | tr '/+' '_-' | tr -d '=')
PASS=$(echo  -s $PASS | openssl dgst -sha256 -binary | base64 | tr '/+' '_-' | tr -d '=')
(echo -e "exp:\t${TS}\nchk:\t${PASTE_LANG}\nlang:\t${PASTE_LANG}"; \
        [ "$PASTE_BURN" -eq "1" ] && echo -e "burn:\ttrue"; \
        [ "$PASTE_GZIP" -eq "1" ] && echo -e "zip:\ttrue"; \
        echo; \
        cat "$PASTE_FILE" | $GZBIN | openssl aes-256-cbc -md md5 -e -a -k "$PASS") | tee sent.txt

HASH=$(cat sent.txt | http POST "${PASTE_URL}/paste")

HASH_OK=$(echo "$HASH" | cut -c1-2)

if [ "$HASH_OK" = "OK" ]; then
  HASH=$(echo $HASH | cut -f2 -d' ')
  URL="${PASTE_URL}/#/${HASH}!${PASS}"

  http --ignore-stdin --form PUT "https://paste.dn42.us/s/cover.${ID}.report" \
    url="$URL" \
    secret="$SECRET"

  echo "url: $URL"
  echo -n "shell: curl -s ${PASTE_URL}/api/get/${HASH} | sed '1,/^\$/d' | openssl aes-256-cbc -md md5 -d -a -k ${PASS}"
  [ "$PASTE_GZIP" -eq "1" ] && echo " | gzip -dc" || echo;
  exit
fi

