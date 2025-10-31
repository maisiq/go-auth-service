#!/bin/sh

vault server -dev &

echo $VAULT_TOKEN | vault login -
vault secrets enable transit
vault write -f transit/keys/jwt-key type=ecdsa-p256

wait -n