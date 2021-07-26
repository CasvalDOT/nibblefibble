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
  "space_id": "your_space_id"
}
```

Each file rapresent a user, so if you need to create an automation for your two employees, Jane and John, simply
create two files with the informations provided in the example.
```bash
ls ~/.config/nibblefibble

# output
- john.json
- jane.json
```

## LICENSE
Licensed under [MIT](./LICENSE)
