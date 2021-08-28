# NIBBLEFIBBLE

A small script used to auto book a desk in your office using `NIBOL` API.

## USAGE

For work, you must manually create a folder in your `HOMEPATH/.config/nibblefibble`
and add inside a JSON file that contains the references of yourself and your desk. For example:
```json
{
  "identity": "John Doe",
  "token": "the_jwt",
  "desk_id": "your_desk_id",
  "space_id": "your_space_id",
  "excluding_days": [6, 7]
}
```

#### Excluding days
The exclude days option allows you to define a list of days in which to skip the desktop booking process.
To use this option you must provide an array of integers that correspond to the number of the day of the week; for example
`[6,7]` means "ignore saturday and sunday".

NOTE!!

You can also create multiple file with different identities and tokens, if you need to booking for multiple persons

## CONFIGURATION FILE

You can set some params and customizations using the file `~/.config/conf.json`

## NOTIFICATION

Nibblefibble support to send a notification if something wrong occured during the desktop rent process.
The notiifcation channel used is Slack (For now). You can configure the notifications using the `conf.json` file.
Here the basic structure for a simple notification.
```json
{
  "slack_hook": "https://hooks.slack.com/services/YOUR/WEB/HOOK",
  "slack_template": {
    "blocks": [
      {
        "type": "header",
        "text": {
          "type": "plain_text",
          "text": "NibbleFibble error",
          "emoji": true
        }
      },
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "{{.Identity}}, an error occured during the rent of your desk"
        }
      }
    ]
  }
}
```

## BUILD
Simply RUN:
```bash
go build
# It generate the executable to launch with the name
# of the project
```

Each file rapresent a user, so if you need to create an automation for your two employees, Jane and John, simply
create two files with the informations provided in the example.
```bash
ls ~/.config/nibblefibble

# output
- john.json
- jane.json
```

## USE WITH CRON
If your system have `cron` service installed. You simply must create a valid cron rule to execute nibblefibble.

## LICENSE
Licensed under [MIT](./LICENSE)
