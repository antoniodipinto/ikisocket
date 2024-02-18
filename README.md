# Last update
## Here's What's Happening

This repository has moved to a new URL, directly under [Fiber's Contrib Repo](https://github.com/gofiber/contrib/). 
This move is part of my effort to enhance the project and offer you better features and support.

**Here**: [https://github.com/gofiber/contrib/blob/main/socketio/README.md](https://github.com/gofiber/contrib/blob/main/socketio/README.md)

### Important note, the name has now changed from ikisocket to Socket.IO

If you wish to switch directly to the new repository, ensure you update your code accordingly.

**From this:**

```go
import (
    "github.com/antoniodipinto/ikisocket"
    "github.com/gofiber/contrib/websocket"
    "github.com/gofiber/fiber/v2"
)

ikisocket.On("EVENT_NAME", func(ep *ikisocket.EventPayload) {})

ikisocket.New(func(kws *ikisocket.Websocket){})
```

**To this**
```go
import (
    "github.com/gofiber/contrib/socketio"
    "github.com/gofiber/contrib/websocket"
    "github.com/gofiber/fiber/v2"
)

socketio.On("EVENT_NAME", func(ep *socketio.EventPayload) {})

socketio.New(func(kws *socketio.Websocket){})
```
## Quick Notes:

- **Bookmark the new repo**: Make sure to star the new repository to keep up with updates.
- **Future Contributions**: Please direct all new issues and contributions to the new repo.
- **Continuous Support**: Your feedback and contributions are always highly appreciated.

---


### WebSocket wrapper for [Fiber v2](https://github.com/gofiber/fiber) with events support
### Based on [Fiber Websocket](https://github.com/gofiber/websocket) and inspired by [Socket.io](https://github.com/socketio/socket.io)

## Star History
[![Star History Chart](https://api.star-history.com/svg?repos=antoniodipinto/ikisocket&type=Date)](https://star-history.com/#antoniodipinto/ikisocket&Date)


