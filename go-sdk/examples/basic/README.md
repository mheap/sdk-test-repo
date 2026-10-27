# SDK Test Example

Run these commands from `go-sdk`. Set `SDK_BASE_URL` to your test API address,
including its path prefix:

```sh
export SDK_BASE_URL=http://localhost:8080/v1
go run ./examples/basic -action bootstrap
go run ./examples/basic
go run ./examples/basic -action cleanup
```

You can use `-base-url` instead of `SDK_BASE_URL`.

`bootstrap.go` creates two disabled destinations, one webhook and one email,
and a subscription for each. The event is `sdk-example.test`. The sample
addresses use `example.com`; they are not real delivery targets. Bootstrap
prints the current destinations and subscriptions as JSON after creation.

`cleanup.go` deletes only destinations with the exact sample names
(`sdk-example-webhook` and `sdk-example-email`) and the sample description
marker. It first deletes subscriptions that link to those destinations and
use the sample event. It does not use forced deletion or delete other
subscriptions. Reserve these names and this marker for this example.

Bootstrap stops if either sample destination name already exists. If a run
fails after it creates some data, run cleanup before you try bootstrap again.
You can run cleanup more than once. Do not change the sample names or marker
before cleanup, or cleanup will not identify that data.

Use a test API, not a production API.
