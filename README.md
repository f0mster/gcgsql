# gcgsql

Golang Code Genaration from SQL

## How does it works

Programm reads yaml file with queries and generate functions with arguments that you need to send to your query. 
If your query is "select" query and for example query name is getUsers - function will return ```[]*getUserStruct``` 
with all fields that you have in query result

todo: return last inserted id

## Usage

1) create yaml file

```
functionname: querystring
...
```

e.g.
```
getUserById: select id:int, name:sql.NullString from users where id=$UserId:int
deleteUserById: delete from users where id=$UserId:int
```

after '$' you must write name of variable and after ':' type of that variable
in select you should 

2) run programm with flags:
	dbtype - type of your database
	yaml - path to yaml file
	package - package of output file
	output - path to generated file

like this:
gcgsql -dbtype=mysql -yaml=./example/mysql/sql.yaml -package=main -output=./example/mysql/out.go