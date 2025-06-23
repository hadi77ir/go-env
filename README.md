# Environment Variable Substitution in Go

This package provides a Go function, `ExpandEnv`, to replace environment variable placeholders in a string with their corresponding values. It supports multiple placeholder formats inspired by shell parameter expansion, using only the standard library (`strings` and `os`) without regular expressions for efficient and readable processing.

## Capabilities

The `ExpandEnv` function:
- Replaces placeholders with environment variable values or specified defaults/errors.
- Uses `strings.Builder` for efficient string construction, minimizing memory allocations.
- Processes placeholders iteratively, handling both simple and complex forms.
- Replaces unset variables in basic forms (`$VAR` and `${var}`) with empty string.
- Supports environment variable setting for the `${var:=word}` form.
- Handles edge cases like malformed placeholders, lone `$`, or special characters gracefully.

## Supported Placeholder Formats

The function supports the following placeholder formats in the input string:

1. **`$VAR`**:
    - Replaces with the value of the environment variable `VAR`.
    - The variable name ends at special characters (`*`, `#`, `$`, `@`, `!`, `?`, `-`, `0-9`) or any non-alphanumeric character (except `_`).
    - If `VAR` is unset, the placeholder (e.g., `$VAR`) is retained.
    - Example: `$USER_NAME` → `Alice` if `USER_NAME=Alice`; `$UNSET` → `$UNSET` if unset.

2. **`${var}`**:
    - Replaces with the value of the environment variable `var`.
    - If `var` is unset, the placeholder (e.g., `${var}`) is retained.
    - Example: `${HOST}` → `localhost` if `HOST=localhost`; `${NO_VAR}` → `${NO_VAR}` if unset.

3. **`${var:-word}`**:
    - Replaces with the value of `var` if set and non-empty; otherwise, uses `word`.
    - Does not modify the environment variable.
    - Example: `${APP_ENV:-dev}` → `dev` if `APP_ENV` is unset or empty.

4. **`${var:=word}`**:
    - Replaces with the value of `var` if set and non-empty; otherwise, uses `word` and sets `var` to `word` using `os.Setenv`.
    - Example: `${UNSET_VAR:=default}` → `default` and sets `UNSET_VAR=default` if unset.

5. **`${var:?message}`**:
    - Replaces with the value of `var` if set and non-empty; otherwise, returns an error.
    - Example: `${NO_VAR:?not set}` → `error: NO_VAR: not set` if `NO_VAR` is unset.

6. **`${var:+word}`**:
    - Replaces with `word` if `var` is set and non-empty; otherwise, uses an empty string.
    - Example: `${USER_NAME:+bob}` → `bob` if `USER_NAME=Alice`; `${NO_VAR:+bob}` → `` if unset.

## Usage

```go
package main

import (
   "fmt"
   "log"
   "os"

   "github.com/hadi77ir/go-env"
)

func main() {
   // Set environment variables for testing
   os.Setenv("USER_NAME", "Alice")
   os.Setenv("APP_ENV", "")

   // Input string with placeholders
   input := "User: $USER_NAME, Env: ${APP_ENV:-dev}, Set: ${UNSET_VAR:=default}, Alt: ${USER_NAME:+bob}, Host: $HOST*test"
   result, err := env.ExpandEnv(input)
   if err != nil {
      log.Fatal(err)
      return
   }
   fmt.Println(result)

   problematicInput := "Err: ${NO_VAR:?not set}"
   result, err = env.ExpandEnv(problematicInput)
   if err != nil {
      log.Fatal(err)
      return
   }
   fmt.Println(result)
}
```

Output:
```
User: Alice, Env: dev, Set: default, Alt: bob, Host: *test
2025/06/24 02:04:04 variable 'NO_VAR' is unset or empty: not set
exit status 1
```

## Notes

- **Variable Names**: For `$VAR`, valid characters are letters, digits, and `_`. The name ends at special characters (`*`, `#`, `$`, `@`, `!`, `?`, `-`, `0-9`) or other non-alphanumeric characters.
- **Performance**: Uses `strings.Builder` with pre-allocated capacity for efficiency.
- **Edge Cases**: Handles lone `$`, malformed placeholders (e.g., `${var`), and special character suffixes (e.g., `$VAR*`) correctly.

## License

MIT License. See [LICENSE](LICENSE) for details.