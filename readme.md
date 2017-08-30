
### /topic/{topic}
- PUT
- payload: json object {}
- response: empty

### /topics
- GET
- payload: none
- response: json array with stored topics


### /subscribe/{topic}/{offset}
- PUT
- payload: json object:
{url: location}

#### params
cmd
- 'follow' gives all entries from the given offset


### /topic/{topic}/{offset}
- GET
- payload: none
- response: json with one log entry 