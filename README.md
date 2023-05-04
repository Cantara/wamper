# wamper
A tool to screenshot websites and post to slack

Using this simple tool to test and demonstrate some simple properties of gober a event driven framework.

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
