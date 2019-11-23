# Triplink
This is a client which parses the logs from [Tripwire](https://github.com/JojiiOfficial/Tripwire) and uploads the IPs to a [server](https://github.com/JojiiOfficial/ScanBanServer). The logs from [Tripwire](https://github.com/JojiiOfficial/Tripwire) are in most cases webscanner who scan your machine(s) to make analytics or want to hack you. If you use this package, you can store scanner automatically in a database and block specific IP addresses. This allows you to sync those evil IPs between multiple devices/servers. In addition you can easily create/restore iptable and ipset backups.

# Install
Run
```
go get
go build -o triplink
```
you can move the binary into /usr/bin if you want:
```
sudo mv ./triplink /usr/bin/triplink
```

# Usage

<b>Create a config file</b> to store the data. Every report/update will go to the given server.<br>
<b>Note:</b> Don't use the same config file for multiple reporter instances
```
# triplink createConfig -f /var/log/Tripwire21 -t <token> -r <https://a-serv.er>
```
<br>
<b>(Report)</b> Parse the logfile and send the new scanner/spammer/hacker IPs to the server. Afterwards update the changed IPs from the server and block them (-u)<br>

```
# triplink report -u
```

<br>
<b>Fetch all IPs</b> from the server and create automatically a set of IPs and blocks them. You can use this command once for getting all ips (existing IPs will be overwritten). If you run this command in eg. a cronjob you can remove the -a it will automatically update new IPs without fetiching everything. Afterwards it will backup and save the IPset<br>

```
# triplink update -a
```

<br>
<b>Backup</b> your <b>IPtables</b> (-t) and IPset (-s) config. Without arguments it will only backup the IPset data. You can turn this off using -s=false<br>

```
# triplink backup -t -s
```

<br>
<b>Restore</b> your <b>IPtables</b> (-t) and IPset (-s) config. Without arguments it will only restore the IPset data. You can turn this off using -s=false. Use it for example in a cronjob with @reboot to restore the IPset data after a reboot, because otherwise they will be lost<br>

```
# triplink restore -t -s
```

<br>
<b>Install</b> one or multiple cronjob(s) to automate reports, fetches, backups and restores<br>

```
# triplink install
```
<b>Note:</b> In some cron installations the $PATH var is not set to the path where iptables or ipset is installed in. If you get an error or the cronjob doesn't work you can either create a symbolic link in `/bin/iptables -> 'your iptables binary'` and `/bin/ipset -> 'your ipset binary` or you can set a custom $PATH in the crontab:
```
PATH=/usr/sbin:/bin:/sbin:/usr/bin      #Make sure ipset and iptables are in one of those folders
```
To uninstall those automations use `crontab -e` and remove the line you don't want to have automated
