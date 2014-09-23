#!/bin/sh
mysql -u testuser -ppassword -f go_lang_test -e "source create_table.sql;"
