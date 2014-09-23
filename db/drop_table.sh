#!/bin/sh
mysql -u testuser -ppassword -f go_lang_test -e "source drop_table.sql;"
