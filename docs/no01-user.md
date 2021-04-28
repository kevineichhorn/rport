# Users
Rport users are provided from JSON file or DB as described in [authentication section](https://github.com/cloudradar-monitoring/rport/blob/master/docs/no03-client-auth.md).

You can manage users with the [REST API](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/cloudradar-monitoring/rport/master/api-doc.yml#/Users).

## API Limitations
Before using the [User Management API](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/cloudradar-monitoring/rport/master/api-doc.yml#/Users), you should provide at least one user belonging `Administrators` group either in a JSON file or DB.

If rport is started with static credentials [auth mode](https://github.com/cloudradar-monitoring/rport/blob/master/docs/no03-client-auth.md#using-a-static-credential), user management API won't be usable.

If rport is started with JSON file credentials, changes to the users list won't be refreshed until rport is restarted since there is a [limitation](https://github.com/cloudradar-monitoring/rport/blob/master/docs/no02-api-auth.md#user-file).

## API Usage
The `/users` endpoints allow you to create, update, delete and list users and add or remove users to/from groups.

As listed in the API docs Users are defined by `username`

### Create
Passwords will be hashed automatically before adding them to file or database.
```
curl -X POST 'http://localhost:3000/api/v1/users' \
-u admin:foobaz \
-H 'Content-Type: application/json' \
--data-raw '{
    "username": "user1",
    "password": "123456"
    "groups":
    {
        "Users",
        "Administrators"
    }
}'
```
### Update
You can provide any parameters that you want to change. 
```
curl -X PUT 'http://localhost:3000/api/v1/users/user1' \
-u admin:foobaz \
-H 'Content-Type: application/json' \
--data-raw '{
    "password": "1234567"
    "groups":
    {
        "Users"
    }
}'
```
This will change password and remove user from Administrators group. To add user to a new group, you should provide all current user groups + a new one e.g.
```
{
    "groups":
    {
        "Users",
        "Administrators",
        "New Group"
    }
}
```

Please note that all changes to the user affecting credentials will have an immediate effect in most cases disregard if you use JWT or basic password auth (e.g. deletion user from Administrators group), so you should use this API carefully.
If you change a password, user will still be able to login with an old JWT token, so the change will work till the next login.

### List all users
```
curl -s -u admin:foobaz http://localhost:3000/api/v1/users
{
    "data": [
        {
            "username": "admin",
            "groups": [
                "Administrators",
                "Users"
            ]
        },
        {
            "username": "root",
            "groups": [
                "Users"
            ]
        }
    ]
}
```
Because of security consideration, users list won't return hashed passwords.
### Delete
```
curl -u admin:foobaz -X DELETE 'http://localhost:3000/api/v1/users/user1'
```
