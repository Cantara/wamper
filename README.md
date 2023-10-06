# wamper
A tool to screenshot websites and post to slack

Using this simple tool to test and demonstrate some simple properties of gober a event driven framework.

## Config

To get wamper running you need a simple base config of crypto keys for application level storage encryption and a auth key for user authentication.

User authentication is subject to change

```
screenshot.key=
screenshot.service.key=
slack.service.key=
authkey=
```

Other values that can be usefull are the following

- webserver.port=3030
- log.dir="log"
- debug.port=6060
- debug.user=user
- debug.pass=pass
- eventstore.host=localhost

If `eventstore.host` is specified the event engine will use eventstore, otherwize it will use local FS

If the debug options are set then the debug module will be enabled, otherwize it is not and does not affect performance

## Slack

For our use we are using a central post bot in slack, with base permissions for only posting.

### Manifest

``` json
{
    "display_information": {
        "name": "Kimmie",
        "description": "A app for posting from generic cantara projects",
        "background_color": "#2c2d30"
    },
    "features": {
        "bot_user": {
            "display_name": "Kimmie",
            "always_online": false
        }
    },
    "oauth_config": {
        "scopes": {
            "bot": [
                "chat:write",
                "files:write"
            ]
        }
    },
    "settings": {
        "org_deploy_enabled": false,
        "socket_mode_enabled": false,
        "token_rotation_enabled": false
    }
}
```
