# monitoring

This is a collection of monitoring checks, tools, ... which I had to write because I did not find any suitable solutions. Besides this repository I wrote also the following monitoring solutions:

* [Tirion](https://github.com/zimmski/tirion) a complete infrastructure for monitoring applications during benchmarks
* [Glamor](https://github.com/zimmski/glamor) a daemon for monitoring hosts via ICMP echo request (ping)

The repository is structured in the following way:

* **/check** every file holds one specific check and can be used on its own
	- **check_gammu_modem** checks if a gammu modem is available, correctly configured and in general can be used
	- **check_imap_mailbox_count** checks the message count of an IMAP mailbox
	- **check_jenkins_http.pl** checks if a Jenkins instance is correctly running
	- **check_smb** checks if a SMB share is read and writable via cifs.mount
	- **check_svn** checks if a SVN repository is readable and writable
	- **check_ntlm_content.pl** check a website using NTLM as authentication
	- **check_sonarqube_http.pl** check if a SonarQube instance is correctly running
	- **check_traceroute.pl** does a traceroute and optionally checks if certain hosts are in the path
* **/check_mk** every folder holds one check_mk check which consist of one agent and one server script
	- **linux_services** agent + check_mk check to check if all services (queried via `service --status-all`) are in an OK status
	- **qnap** agent + check_mk check for qnap hardware
	- **zypper** agent + check_mk check for openSUSE's zypper to check for new updates
* **/scripts** every folder holds one script for various purposes
	- **cronjob-mailgw-remove-queued-spam** removes mails that were too long in the mail queue and can be marked as spam
	- **rebind-usb-umts-modem.pl** rebinds an UMTS modem which is handy if it often hangs
* **uptime-statistic.daemon.pl** a more detailed `ping` command
