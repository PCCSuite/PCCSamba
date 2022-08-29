# How to deploy

## only first
`docker compose run -it sambad init`

## to disable password comlexity
`docker compose exec sambad samba-tool domain passwordsettings set --complexity=off`

## to disable LDAP strong auth
```
[global]
	ldap server require strong auth = no
```