## Chat room example

###Disclaimer:
__This my personal interpretation of a basic chatroom example, with basic API to handle the rooms. Please feel free to add comments or suggestions here__


### REST API for room management

## Rooms
__CREATE__
- Method: __*POST*__
- Endpoint: [API-URL]/rooms/create
- Description: Create a room via API passing JSON name

__PARAMS__
- name (mandatory)
```json
{
  "name":"Best Room ever"
}
```

__RESPONSE__
```json
{
  "name": "Best Room ever",
  "uuid": "xgYENKyyAdOdrSZb5JNMAu1g7iEGlaE77cEEZgaHg3n5dGWaFG2IV0Nru6C1QEAKh2F9CL4I8uxW1CK5ev0g9mnJTHy0pBSrndGg",
  "users": null // list of users in that room
}
```
---
__GET__
- Method: __*GET*__
- Endpoint: [API-URL]/rooms
- Description: Retrieve all the rooms created


__RESPONSE__
```json
[
  {
    "name": "Best Room ever",
    "uuid": "xgYENKyyAdOdrSZb5JNMAu1g7iEGlaE77cEEZgaHg3n5dGWaFG2IV0Nru6C1QEAKh2F9CL4I8uxW1CK5ev0g9mnJTHy0pBSrndGg",
    "users": null
  }
]
```
---

__DELETE ROOM__
- Method: __*DELETE*__
- Endpoint: [API-URL]/rooms/delete/__[UUID]__
- Description: Delete room by UUID

__PARAMS__
- Room UUID (mandatory)

__RESPONSE__
```json
true
```
---
__JOIN__
- Method: __*POST*__
- Endpoint: [API-URL]/rooms/join
- Description: Join the room passing room id and user id

__PARAMS__
- room (Room UUID | Mandatory)
- user (User ID | Mandatory)
```json
{
  "user":"1",
  "room":"xgYENKyyAdOdrSZb5JNMAu1g7iEGlaE77cEEZgaHg3n5dGWaFG2IV0Nru6C1QEAKh2F9CL4I8uxW1CK5ev0g9mnJTHy0pBSrndGg"
}
```

__RESPONSE__
```json
true|false
```
---


### Connect to the websocket
```
ws://localhost:3000/ws/<user-id>
```


### Message object example

```
{
    "from": "<user-id>",
    "to": "<recipient-user-id>",
    "room": "<room-id>",
    "data": "hello"
}
```
