# Client Authentication
The Rportd can read client auth credentials from three different sources.
1. A "hardcoded" single credentail pair
2. A file with client credentials
3. A database table

Which one you choose is an either-or decision. A mixed-mode is not supported.

If you select option 2 or 3 client access and the needed credentials can be managed through the API and the UI.

## Using a static credential

::: danger
A single static clientauthid-password pair is not recommended for productive use. If all your clients use the same credential you cannot expire clients individually. If the password falls into wrong hands you must reconfigure all your clients.
:::

To use just a single pair consisting of a client-auth-id and a password enter the following line to the server config(`rportd.config`) in the `[server]` section.
```
auth = "rport:a-strong-password12345"
```
Make sure no other auth option is enabled.
Reload rportd to activate the changes.
Quite simple. Now you can run a client using the client-auth-id `rport` and the password `a-strong-password12345`. It can be done in two ways:
1. Use a command arg: `--auth rport:a-strong-password12345`
2. Enter the following line to the client config(`rport.config`) in the `[client]` section.
```
auth = "rport:a-strong-password12345"
```

## Using a file
If you want to have more than one credential, create a json file with the following structure.
```
{
    "clientAuth1": "1234",
    "admin":       "123456",
    "client1":     "yienei5Ch",
    "client2":     "ieRi1Noo2"
}
```
Using `/var/lib/rport/client-auth.json` is a good choice.

Enter the following line to your `rportd.config` in the `[server]` section.
```
auth_file = "/var/lib/rport/client-auth.json"
```
Make sure no other auth option is enabled.
Reload rportd to activate the changes.

The file is read only on start. Changes to the file, while rportd is running, have no effect.

If you want to manage the client authentication through the API make sure the auth file is writable by the rport user for example by executing `chown rport /var/lib/rport/client-auth.json`.

## Using a database table
Clients auth credentials can be read from and written to a database table.

To use the database client authentication you must set up a global database connection in the `[database]` section of `rportd.conf` first.
Only MySQL/MariaDB and SQLite3 are supported at the moment. The [example config](https://github.com/cloudradar-monitoring/rport/blob/master/rportd.example.conf) contains all explanations on how to set up the database connection.

The tables must be created manually.


:::: code-group
::: code-group-item MySQL
```mysql
CREATE TABLE `clients_auth` (
  `id` varchar(100) PRIMARY KEY,
  `password` varchar(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```
:::
::: code-group-item SQLite3
```sqlite
CREATE TABLE `clients_auth` (
  `id` varchar(100) PRIMARY KEY,
  `password` varchar(100) NOT NULL
);
```
:::
::::

Having the database set up, enter the following to the `[server]` section of the `rportd.conf` to specify the table names.
```
auth_table = "clients_auth"
```
Reload rportd to apply all changes.

## Manage client credentials via the API

The [`/clients-auth` endpoint](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/cloudradar-monitoring/rport/master/api-doc.yml#/Rport%20Client%20Auth%20Credentials) allows you to manage clients and credentials through the API.
This option is disabled if you use a single static clientauthid-password pair.
If you want to delegate the management of client auth credentials to a third-party app writing directly to the auth-file or the database, consider turning the endpoint off by activating the following lines in the `rportd.conf`.
```
## If you want to delegate the creation and maintenance to an external tool
## you should turn {auth_write} off.
## The API will reject all writing access to the client auth with HTTP 403.
## Applies only to auth_file and auth_table
## Default: true
auth_write = false
```

List all client auth credentials.

```
curl -s -u admin:foobaz http://localhost:3000/api/v1/clients-auth|jq
{
  "data": [
    {
      "id": "clientAuth1",
      "password": "1234"
    },
    {
      "id": "client1",
      "password": "yienei5Ch"
    },
    {
      "id": "client2",
      "password": "ieRi1Noo2"
    }
  ]
}
```

Add a new client auth credentials

```
curl -X POST 'http://localhost:3000/api/v1/clients-auth' \
-u admin:foobaz \
-H 'Content-Type: application/json' \
--data-raw '{
    "id":"client3",
    "password":"hase243345"
}'
```
