#!/bin/sh
set -e

ensure_dirs ()
{
    for j in "$@"; do
         test -d "$j" || mkdir -p "$j"
    done;
}

check_files_exist () 
{
    for j in "$@"; do
        if [ -f "$j" ]; then
            true
        else
            echo "File $j not found - unable to install"
            exit 1
        fi
    done;
}

echo "Obtaining directory for semdict.service"
SYSTEMD_DIR=`pkg-config systemd --variable=systemdsystemunitdir`

check_files_exist semdict semdict.service semdict.config.json.example templates/general.html

ensure_dirs /etc/semdict /usr/share/semdict/templates

cp semdict /usr/bin
cp semdict.service $SYSTEMD_DIR

CONFIG_FILE=/etc/semdict/semdict.config.json

cp semdict.config.json.example ${CONFIG_FILE}.example
chmod 600 ${CONFIG_FILE}.example

cp -R templates /usr/share/semdict/

systemctl daemon-reload

echo Sample config is ${CONFIG_FILE}.example. To run semdict, 
echo You must provide ${CONFIG_FILE}. Don\'t forget to make it 
echo secure, e.g. with «sudo chmod 600».
echo To run semdict, use «sudo service semdict start»

