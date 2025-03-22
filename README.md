
# Drive API

This project is a basic GO API based on Google Drive. It allows you to authenticate users with ABAC (Attribute-Based Access Control) and JWT tokens. It also uses MongoDB to control user access and to manage the distribution of files and folders.


## Installation

Download this project.
Install the Drive API with go

```bash
  go mod download
  go build cmd/main.go
```
    
Optionally, you can mount a drive in the 'files' folder to improve file management.

```bash
  mount /dev/sdb7 files
```
## API Reference

#### Login

```http
  POST /login
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `username`| `string` | **Required**. Name of the user    |
| `password`| `string` | **Required**. Key of the user     |

##### Result: JWT Token string (Must be used on Authentication header)

#### Logout

```http
  GET /logout
```

##### Result: logout message

#### Get folder

```http
  GET /drive/d/{id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. Id of item to fetch |

##### Result: Array with Contents ID

#### Delete resource

```http
  GET /drive/r/{id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. Id of item to fetch |

##### Result: Delete status message

#### Get file

```http
  GET /drive/f/{id}
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. Id of item to fetch |

##### Result: File binary

#### Upload file

```http
  POST /drive/upload
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `parent`  | `string` | ID of a directory if applies      |
| `file`    | `binary` | **Required**. Data of the file    |

##### Result: ID of the new element

#### Create folder

```http
  POST /drive/create
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `parent`  | `string` | ID of a directory if applies      |
| `name`    | `string` | **Required**. Name of the folder  |

##### Result: ID of the new element




## License

This project is licensed under the [MIT](https://choosealicense.com/licenses/mit/) license, which means you can freely use, modify, and distribute the code, provided you retain the original copyright notice and this same license on any copies or derivative versions.