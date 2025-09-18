#!/usr/bin/env bash

set -euo pipefail

# Request the OIDC token (audience must match your trust policy)
resp=$(curl -sSL "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=sts.amazonaws.com" \
  -H "Authorization: bearer ${ACTIONS_ID_TOKEN_REQUEST_TOKEN}")

jwt=$(echo "$resp" | jq -r '.value')

if [ -z "$jwt" ] || [ "$jwt" = "null" ]; then
  echo "Failed to fetch OIDC token. Do you have 'permissions: id-token: write'?"
  echo "Response was: $resp"
  exit 1
fi

# Decode the JWT payload (base64url)
payload=$(echo "$jwt" | cut -d '.' -f2 | tr '_-' '/+' | base64 -d 2>/dev/null || true)
# If padding was missing, try with padding
if [ -z "$payload" ]; then
  payload=$(echo "$jwt" | cut -d '.' -f2 | tr '_-' '/+' | awk '{ l=length($0)%4; if(l>0) printf "%s", $0 substr("====",1,4-l); else print }' | base64 -d)
fi

echo "OIDC payload:"
echo "$payload" | jq .

sub=$(echo "$payload" | jq -r '.sub')
echo "OIDC sub: $sub"
