### Description:
Fire custom events using sockets and the standard .On() listener
### Connect to the websocket
```
ws://localhost:3000/ws/<user-id>
```
### Message object example

```
{
"from": "<user-id>",
"to": "<recipient-user-id>",
"event": "CUSTOM_EVENT",
"data": "hello"
}
```