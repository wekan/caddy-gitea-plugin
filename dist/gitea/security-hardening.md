# security-hardening

What GitHub code scanning found in the parts of Gitea this plugin builds in,
and was worth changing. Inside Weddy, Gitea shares its process with Caddy and
WeKan, so a request that exhausts memory stops all of them.

- **Container tag list** (`models/packages/container/search.go`): the capacity
  of the result slice was the request's `?n=`, so
  `/v2/<image>/tags/list?n=1000000000` reserved gigabytes before one row was
  read - for a public package, without signing in. The capacity is now at most
  100; `n` still limits the query.
- **32-bit conversions** (`services/context/api.go`,
  `services/migrations/github.go`): an `int64` total and an `int64` issue number
  became `int` without a bound, which wraps on the 32-bit platforms Weddy is
  built for (linux/386, arm). The total is clamped as `newPagination` already
  did, and an issue number outside 1..MaxInt32 is refused.
- **Access log** (`services/context/access_log.go`): the template got the
  request itself, so a token in the query string or a credential header could
  be written to the log. It now gets a copy with `token`, `access_token`,
  `private_token` and `password`, and `Authorization`, `Cookie`, `X-Gitea-OTP`
  and `Proxy-Authorization`, replaced by `[REDACTED]`; the request is unchanged.
  `TestRedactedRequest` pins it.

The other alerts in `gitea/` were reviewed and are not changed here; why is in
the plugin wiki's
[Hall of Fame](https://github.com/wekan/caddy-gitea-plugin/wiki/Hall-of-Fame).
