# TrueWatcher

It's a small utility that uses the [TrueNAS jsonrpc API](https://api.truenas.com/v25.04/) to keep an eye on the TrueNAS
managed applications and update those once they have a new available version.

## What do you need
### URL
The URL that points to your TrueNAS instance, is in a [WebSocket format](https://datatracker.ietf.org/doc/html/rfc6455#section-11.1.1).
Some actual examples are:
* `ws://192.168.0.10/api/current` - if your TrueNAS is configured to allow API calls on the non-ssl port, this is what you need.
  * If you use this against an instance that does not support non-ssl port, the API token used will be revoked right away. You will have to rotate it.
* `wss://192.168.0.10/api/current` - if your TrueNAS allows traffic on the ssl port, this is the way to go.
  * This is the way to go even if you don't have a valid SSL certificate for your instance.

The examples above refer to `non-ssl port` as 80 and `ssl port` as 443. 
If you changed your default ports (ie: System -> General Settings -> GUI Settings -> Web Interface HTTP Port && Web Interface HTTPS Port),
then you'll have to add the port in the examples above.
As an example, if you changed the 80 port to 8080 and 443 to 8443, your url should look like the following:
* `ws://192.168.0.10:8080/api/current`
* `wss://192.168.0.10:8443/api/current`

### User API token
To make this work, it's recommended to use a User API Token.
This can be accessed by navigating to Credentials -> Users -> Api Keys (top-right corner).

The recommended way to do this, is to create a specific user (called service account) only for this specific purpose.
This user will have a group that will have a list of limited privileges, strictly for this job.
To do so, follow the steps:
* Create the privilege:
  * Credentials -> Groups -> Privileges -> Add:
    * Give it a name
    * Select the following privileges: 
      * Apps Read
      * Apps Write
* Create the system group:
  * Credentials -> Groups -> Add
    * Give it a name
    * Select your newly created privilege
    * Save
* Create the user:
  * Credentials -> Users -> Add
    * Give it a name and a username
    * Disable password
    * Disable "Create new primary group" and instead select your newly created group in the "Primary group"
    * Disable "SMB User"
* Create the API Key:
  * Credentials -> Users -> Api Keys -> Add
    * Give it a name
    * Select your newly created service account
    * Set an expiration date (optional and if you enable it, will require you to come in this screen again and rotate it)
    * Click "Save" and copy the given API Key. Put it in a safe spot for the moment since you'll need it later

The easiest, but not recommended way to do this, is by issuing a token for your `truenas_admin` user.

### Run the application
Just export the two environment variables and run the application:
```shell
export TRUENAS_URL="<your url>"
export TRUENAS_API_KEY="<the token you copied above>"
```

### Other available configurations
* Delay between checks (env var: `CHECK_DELAY`; default: "6h")
  * Control how often the applications are checked for available updates.
    It gets a **duration** in [golang format](https://pkg.go.dev/time#ParseDuration) (eg: "1h", "5s", "2h45m", etc).
* Filtering by application name
  * Whitelisting (env var: `APP_WHITELIST`; default: "")
    * Updates only the specified applications. When not specified, filtering for this is disabled.
      The value is a comma separated string (eg: "Portainer,SeaweedFS,immich").
      All the names are lower cased and checked accordingly.
  * Blacklisting (env var: `APP_BLACKLIST`; default: "")
    * Does not update the specified application names. When not specified, filtering for this is disabled.
      The value is a comma separated string (eg: "Portainer ,SeaweedFS,immich").
      All the names are lower cased and checked accordingly.
  * Both options can be used together, and the first one that excludes an application will filter it out.
  * None can be used which will allow upgrading any application that has an upgrade available.

### Use the provided docker image
In case you want to run this in your Portainer or Dockage instance, you can use the already existing and up to date [docker image](https://hub.docker.com/r/yottta/truewatcher).

> The timezone of the container can be configured by using the TZ environment variable.
