# Cynosure

An orchestrator for managing local applications, remotely via API.

## Getting cynosure

Cynosure is a standard golang executable and can be installed from source via:

```bash
go install github.com/norganna/cynosure
```

## Starting cynosure

Simply run the server and it will start up automatically and generate a starter configuration for you at `~/.cyno/config` if you don't already have one.

```bash
cynosure server
```

## Images

Cynosure runs processes in a `chroot`-ed environment for you, based off files provided within a `.tar.gz` "image" file.

The contents of the image file are simply the contents of the root filesystem you want your process to have.

When cynosure launches your process it will extract your files into a temporary directory for you before executing your entry binary with a chroot into that directory.

## Starting a process

To start a process you call the `Start` API and give it a `StartRequest` (see below for full definition) that defines the process.

An example of a simple process would look like:
```javascript
{
    command: {
        name: "ping",
        image: "ping:v1",
        entry: "/bin/ping",
        args: ["google.com"]
    }
}
```

This assumes there is an image at `${root}/images/ping/v1.tar.gz` (`${root}` is defined in your `~/.cyno/config`) that contains a `/bin/ping` executable suitable for your OS/ARCH.

You can create a suitable file simply by doing:

```bash
mkdir ~/.cyno/root/images/ping
tar cvzf ~/.cyno/root/images/ping/v1.tar.gz /sbin/ping
```

Try it now using `cynosure curl` to hit the http API:

```bash
cynosure curl -X POST \
  https://localhost:8055/v1/start \
  --data-raw \
    '{"command":{"name":"ping","image":"ping:v1","entry":"/sbin/ping","args":["google.com"]}}'
```

This is the same as using the normal curl command, but cynosure automatically generates some certificates for you authenticate with.
(You *could* easily generate these certs yourself using the `cynosure config client-cert` command and run curl yourself.)

## More sophisticated usage

Obviously cynosure isn't meant for operation by hand, it provides an API for you to manage the entire process remotely, from updating images, setting up environments, stopping existing processes, viewing output logs, etc.

You can view the full API definition either in protobuf format (at [cyno.proto](./proto/cynosure/cyno.proto)) or swagger (at [cyno.swagger.json](./proto/cynosure/cyno.swagger.json)).

You can either write your applications or scripts directly using the HTTP API, using the gRPC API, or using cynosure itself on the remote machine.

To generate a configuration which you can use remotely, you can run:

```bash
cynosure config > cynosure.conf

# OR, to just to generate the PEM certificates:
cynosure config client-cert 
```

This will generate a cynosure compatible client config file (OR certificates for you to use) to STDOUT. You can copy this file to your remote machine.

You can then write your application or script, or install and run cynosure on the remote machine to connect back to the server using the client config file.

```bash
cynosure --config=cynosure.conf curl http://REMOTEIP:8055/v1/running
```

## StartRequest definition

A `StartRequest` is structured like this:

```
// StartRequest is the input supplied to the `Start` API endpoint.
message StartRequest {
    // Command to run (or that is running)
    Command command {
        // Name will be used as the prefix for the identifier.
        string name
        
        // Image is the uploaded image file to use as the filesystem.
        string image
        
        // Entry is the command to execute as the entry-point.
        string entry
        
        // Args are supplied to the executable.
        string[] args
        
        // Env provides extra environment variables.
        string[] env
        
        // Requirements is a set of dependencies to be met before the command will be run.
        map<string, Deps> requirements {
            // Deps is the list of `Dep` items.
            Dep[] deps {
                // Identity of the broker to use, defined within the server configuration.
                string identity
                
                // Wait is the thing to wait for, how it is interpreted/found is up to the provider.
                string wait
            }
        }
    }

    // Namespace to run the command in.
    string namespace

    // Labels to assign to the process (allows filtering of processes).
    KV[] labels {
        // Key of the value.
        string key

        // Value of the key.
        string value
    }
    
    // Environments to run the command in (stacks environment variables).
    string[] environments
    
    // Watches allow observation of key log entries and changing the process ready state.
    map<string, Watch> watches {
        // Match is a string to find in the output that triggers this watch.
        string match

        // State determines whether this match will make the app ready, not, or do nothing.
        State state {
            // Unchanged does not change the state of the process (default).
            Unchanged
            // MakeReady changes the process state to ready, if not currently not-ready.
            MakeReady
            // NotReady changes the process state to not-ready, if currently ready.
            NotReady
        }
    }
}
```



# TODOs

* Lots of things...
* Finish off the etcd, consul and kube providers.
* Add external cyno support to the cyno provider.
