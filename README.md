# Tripwire-reporter
This is a client which parses the logs from [Tripwire](https://github.com/JojiiOfficial/Tripwire) and uploads the ips to a server. The logs form [Tripwire](https://github.com/JojiiOfficial/Tripwire) are in most cases webscanner who scan you or your server to make analytics or want to hack you. If you use this package, you can store scanner in a database automatically and block specific ip adderesses. This allowes you to sync those evil ips between multiple devices/server. In addition you can easily create/restore iptable and ipset backups.

# Install
Run
```
go get
go build -o twreporter
```
you can move the binary into /usr/bin if you want using
```
sudo mv ./twreporter /usr/bin/twreporter
```

# Usage

Create a config file to store the data. Every report/update will go to the given server.<br>
```
# twreporter createConfig -f /var/log/Tripwire21 -t <token> -r <https://a-serv.er>
```
<br>
Parse the logfile and send the new scanner/spammer/hacker IPs to the server. Afterwards update the changed IPs from the server and block them (-u)<br>

```
# twreporter report -u
```

<br>
Fetch all IPs from the server and create automatically a set of IPs and blocks them. You can use this command once for getting all ips (existing IPs will be overwritten). If you run this command in eg. a cronjob you can remove the -a it will automatically update new IPs without fetiching everything. Afterwards it will backup and save the IPset<br>

```
# twreporter update -a
```

<br>
Backup your IPtables (-t) and IPset (-s) config. Without arguments it will only backup the IPset data. You can turn this off using -s=false<br>

```
# twreporter backup -t -s
```

<br>
Restore your IPtables (-t) and IPset (-s) config. Without arguments it will only restore the IPset data. You can turn this off using -s=false. Use it for example in a cronjob with @reboot to restore the IPset data after a reboot, because otherwise they will be lost<br>

```
# twreporter restore -t -s
```

