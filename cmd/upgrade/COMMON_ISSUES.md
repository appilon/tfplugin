# TROUBLSHOOTING

## Upgrading to terraform 0.12
I have noticed for some providers this leads to a compilation error in a transitive dependency? It might be worth specifying `-u` which will upgrade transitive dependencies to their latest `MINOR` release, this might solve the issue however in my specific experiences it fails without a helpful error message.

```
$ GO111MODULE=on go get -u github.com/hashicorp/terraform@pluginsdk-v0.12-early2
```

UPDATE
Digging deeper there appears to be a problem with a transitive dependency `cloud.google.com/go`. This is relied on by many projects, notably grpc which itself is relied on by a lot of this. you can inspect a providers graph with

```
$ GO111MODULE=on go mod graph
```

If you encounter an error when upgrading to the new sdk regarding that dependency, try this.
```
$ GO111MODULE=on go get -u cloud.google.com/go@master
$ GO111MODULE=on go get github.com/hashicorp/terraform@pluginsdk-v0.12-early2
$ GO111MODULE=on go mod tidy
$ GO111MODULE=on go mod vendor
```