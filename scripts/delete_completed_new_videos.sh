#!/bin/sh

if [ "$MYSQL_HOGESUKE_PASS" = "" ]; then
    echo "Please set MYSQL_HOGESUKE_PASS to the environment variable."
    exit 1
fi

mysql -u hogesuke -p$MYSQL_HOGESUKE_PASS nicotune -e "DELETE FROM new_videos WHERE status = 1;"
