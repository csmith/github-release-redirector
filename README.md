# github-release-redirector

Small go application that listens for incoming HTTP requests and redirects
to assets from the latest release published to a GitHub repository.

## Usage

```console
$ go install .
$ github-release-redirector -port 1234 -repo user/repo
```

or using docker:

```console
$ docker run -p 8080:8080 ghcr.io/csmith/github-release-redirector -repo user/repo
```

## Options

```
Usage of github-release-redirector:
  -poll int
    	the amount of time to wait between polling for releases; 0 to disable polling (default 3600)
  -port int
    	the port to listen on for HTTP requests (default 8080)
  -redirect string
    	if specified, requests for / will be redirected to this url
  -repo string
    	the repository to redirect releases for, in user/repo format [required]
  -webhook string
    	full path to receive release webhooks from GitHub on
```

## Notes

### Root redirecting

If the `-redirect` option is specified, then requests to the root (i.e. `/`)
will be redirected to that URL, and no further processing will be done for
that request.

### Polling and webhooks

Releases are refreshed at startup and then, by default, once an hour. You can
customise this duration with the `-poll` option; setting the poll time to `0`
will disable polling entirely other than at startup.

As an alternative to polling, you can configure a webhook URL with the
`-webhook` argument. This must be the full URL that GitHub will call
when a release is made, and it is advisable to include a secret in the
URL to avoid any client on the Internet being able to trigger a refresh.

For example if you run with `-webhook /webhook/gjrCBVy7`, you
should configure GitHub to send pull requests to
`https://yourserver.example.com/webook/gjrCBVy7`. You should configure
the WebHook to only send individual events, and select only the
`Releases` option.

The contents of the webhook are not used; it is merely used as a signal
to start a refresh using the GitHub API.

When running with the `-webhook` option it is recommended to also disable
polling with `-poll 0`, or set the polling time to a much larger value.

### Asset matching

Assets are matched on their name, so an asset attached to the latest release
named "myproject-installer-1.2.3.exe" will be available at the URL
`/myproject-installer-1.2.3.exe`.

Any url not mapped to an asset is responded to with a 404 not found error;
if there is no latest release available all requests will respond with 500
internal server error.

### HTTPS

The service listens only on HTTP, not HTTPS. For production use it should
be placed behind an SSL-terminating proxy such as Traefik, Nginx or HAProxy.

## Contributing/further work

This performs the job I wrote it to do, so it's unlikely I'll be developing
any further features. If you're using this and would like to send a pull
request for a feature I'd be happy to review and merge.

In particular I'd welcome the following:

- [ ] HTTPS support (specifying certs/keys/etc)
- [ ] Support for GitHub's secret digests, instead of using a secret URL
- [ ] Parsing of the WebHook content to avoid refreshing needlessly

## Licence

This software is released under the MIT licence. See the LICENCE file for
full details.
