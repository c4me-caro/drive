
# Drive API

This project is a basic GO API based on Google Drive. It allows you to authenticate users with ABAC (Attribute-Based Access Control) and JWT tokens. It also uses MongoDB to control user access and to manage the distribution of files and folders.



## Installation

Download this project.
Install the Drive API with go

```bash
  go mod tidy
  go build cmd/main.go
```

Optionally, you can mount a drive in the 'files' folder to improve file management.

```bash
  mount /dev/sdb7 files
```



## Mount systemd service:

Update this variables on the `drive-api.service` both with absolute path:
- `EXecStart`: your go generated binary
- `WorkingDirectory`: path of your application directory

Run the following commands

```bash
  cp drive-api.service /etc/systemd/system/drive-api.service
  systemctl enable --now drive-api.service
```



## API Reference

#### Login

```http
  POST /login
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `username` | `string` | **Required**. Name of the user    |
| `password` | `string` | **Required**. Key of the user     |

##### Result: JWT Token string (Must be used on Authentication header)


#### Logout

```http
  GET /logout
```

##### Result: logout message


#### Check User

```http
  GET /validateUser
```

##### Result: JWT token or error


#### Get service API Key

```http
  GET /newApiKey
```

##### Result: API Key string


#### Get file

```http
  GET /drive/f/{id}
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required**. Id of item to fetch |

##### Result: File binary


#### Get folder

```http
  GET /drive/d/{id}
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required**. Id of item to fetch |

##### Result: requested resource


#### Delete file

```http
  GET /drive/r/{id}
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required**. Id of item to fetch |

##### Result: Delete status message


#### Delete folder

```http
  GET /drive/rd/{id}?recursive=false
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `id`       | `string` | **Required**. Id of item to fetch |
| `recursive`| `string` | Delete childrens. true or false   |

##### Result: Delete status message


#### Upload file

```http
  POST /drive/upload/{parent}
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `parent`   | `string` | ID of a directory if applies      |
| `file`     | `binary` | **Required**. Data of the file    |

##### Result: created resource


#### Create folder

```http
  POST /drive/create/{name}
```

| Parameter  | Type     | Description                       |
| :--------  | :------- | :-------------------------------- |
| `parent`   | `resobj` | json body of the parent wanted    |
| `name`     | `string` | **Required**. Name of the folder  |

##### Result: created resource



## License

This project is licensed under the [MIT](https://choosealicense.com/licenses/mit/) license, which means you can freely use, modify, and distribute the code, provided you retain the original copyright notice and this same license on any copies or derivative versions.
