---
title: "kf map-route"
slug: kf-map-route
url: /docs/general-info/kf-cli/commands/kf-map-route/
---
## kf map-route

Map a route to an app

### Synopsis

Map a route to an app

```
kf map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH] [flags]
```

### Examples

```

  kf map-route myapp example.com --hostname myapp # myapp.example.com
  kf map-route --namespace myspace myapp example.com --hostname myapp # myapp.example.com
  kf map-route myapp example.com --hostname myapp --path /mypath # myapp.example.com/mypath
  
```

### Options

```
  -h, --help              help for map-route
      --hostname string   Hostname for the route
      --path string       URL Path for the route
```

### Options inherited from parent commands

```
      --config string       config file (default is $HOME/.kf)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --namespace string    kubernetes namespace
```

### SEE ALSO

* [kf](/docs/general-info/kf-cli/commands/kf/)	 - kf is like cf for Knative
