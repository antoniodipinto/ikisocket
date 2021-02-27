### Disclaimer:
__This my personal interpretation of a basic chatroom example, with basic API to handle the rooms. Please feel free to add comments or suggestions here__


### Connect to the websocket
```
ws://localhost:3000/ws/<user-id>
```
### Message object example

```
{
"from": "<user-id>",
"to": "<recipient-user-id>",
"data": "hello"
}
```