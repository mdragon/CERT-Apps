echo 1 > out.txt
start out.txt
dev_appserver.py .\ -d --use_sqlite --port_sqlite_data 1> out.txt 2>&1