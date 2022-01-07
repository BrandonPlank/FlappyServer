systemctl stop flappyserver.service
runuser -l api -c 'cd /home/api/FlappyServer; git reset --hard; git pull; chmod +x FlappyServer'
systemctl start flappyserver.service