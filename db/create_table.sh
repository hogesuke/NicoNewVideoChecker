#!/bin/sh
mysql -u testuser -ppassword -f go_lang_test -e "source create_table.sql;"
mysql -u testuser -ppassword -f go_lang_test -e "source create_index.sql;"
