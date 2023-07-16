# dns-go
A toy dns-resolver written in go based on https://implement-dns.wizardzines.com/ by Julia Evans

```
go run dns.go twitter.com
```

```
Querying 198.41.0.4 for twitter.com
Querying 192.12.94.30 for twitter.com
Querying 198.41.0.4 for a.r06.twtrdns.net
Querying 192.12.94.30 for a.r06.twtrdns.net
Querying 205.251.195.207 for a.r06.twtrdns.net
Querying 205.251.192.179 for twitter.com
The IP of twitter.com is 104.244.42.193
```