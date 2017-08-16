# repeater
Simple repeater logic for Go

Small go library for repeating a fallible operation until it succeeds. Example use:

```go
import (
  "github.com/fixermark/repeater"
)

func SomeUnreliableFunction() error {
   . . .
}

func main() {
  r := repeater.Default()
  err := r.Repeat(SomeUnreliableFunction)
  
  // if SomeUnreliableFunction returns error,
  // r retries it with exponential backoff delay.
  // If it fails ten times, r.Repeat will return
  // the most recent error.
}
```

## Features

* DefaultInfinite() provides infinite retries
* Customize your own repetition policy with NewRepeater
