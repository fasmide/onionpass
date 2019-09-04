onionpass - connect to hidden services with ease instead of privacy
---
onionpass is about the convenience of being able to access hidden ssh services by just passing `-J onionpass.net` to regular ssh commands.

onionpass accepts any combination of username and password/key.

`onionpass.net` will be available until it runs out of bandwidth or file descriptors - use at your risk.

Usage examples
```
$ ssh -J onionpass.net root@bs....geq.onion
```
With ssh's `-J` parameter.

```
$ scp -o ProxyJump=onionpass.net some_file.txt bs....geq.onion:
```
Using SCP is a little harder as it has no `-J` parameter, one could add the setting to a special host in your `.ssh/config` file as shown below.

```
$ cat .ssh/config
Host my_hidden_ssh_daemon
    HostName bs....geq.onion
    User john
    ProxyJump onionpass.net
$ ssh my_hidden_ssh_daemon
```
Using onionpass in a ssh_config.

```
ssh onionpass.net -N -L8080:bs....geq.onion:80
```
Browse an onion web server by visiting localhost:8080. 
onionpass accepts any forward ssh connection, but only to `.onion` addresses.