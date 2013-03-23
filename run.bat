echo 1 > out.txt
start out.txt
dev_appserver.py .\ --log_level debug --dev_appserver_log_level debug 1> out.txt 2>&1