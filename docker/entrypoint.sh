if [ "$1" = "init" ]; then
  exec samba-tool domain provision --use-rfc2307 --interactive
fi
if [ ! -f "/etc/samba/smb.conf" ]; then
    echo "init samba first"
    exit 1
fi
# ionice -c 3 smbd -F --no-process-group </dev/null &
ionice -c 3 samba -i </dev/null &
exec /usr/bin/SambaAPI