# gcgsql

Golang Code Generation from SQL

## How does it works

Program reads yaml file with queries and generates functions with arguments that you need to pass to your query. 
If your query is "select" query and for example function name is getUsers it will return ```[]*getUserStruct``` 
with all fields that you have in query result

## Yaml

Create yaml file like this:

```
package: "main"
sqlDriver: "github.com/go-sql-driver/mysql"
outFilePath: "out.go"
withContext: true
withTransaction: true
imports:
  - "github.com/someauthor/somepkg2"
  - "github.com/somethingelse/somethingels"
queries:
  getAllUsers: "select id:int, name:string from users"
  findUsersWith: "select id:int, name:string, info:string, login:string from users where id=$userId:int"
  getUserByName: "select name:string, id:int from users where name like $name:string or name like $altername:string"
  updateUserName: "update users set name=$name:string where id=$userId:int"
  addUser: "insert into users set name=$name:string"
  findUsersIdIn: "select id:int, name:string, info:string, login:string from users where id in ($userId:[]int)"
  addbulkusers: "insert into users (id, name) values $userdata#($id:int, $name:string)|,#"
  addbulkusers2: "insert into users (id, name) values $userdata:sometype#($id, $name)|,#"
  findUsersIdInOrByLoginAndName: "select id:int, name:string, info:sql.NullString, login:sql.NullString from users where id in ($userId:[]int) $ln# or (login=$login:string and name=$name:string)|#  or admin=1"
```

Where:

**outFilePath** - path and filename of a generated file  
**package** - golang package of an output directory  
**sqlDriver** - sql drivers have some differences when working with arguments, so you should specify the driver.  
	for now, we support:  
		1) github.com/go-sql-driver/mysql for mysql  
		2) github.com/lib/pq for postgres  
**imports** - if you need some additional type support you should specify them in this property  
**queries** - this is a map, where key is the name of the generated function and value is the query.  
**withContext** - bool. should Context support be added to generated functions
**withTransaction** - bool. should Transaction support be added to generated functions

## Syntax
### Selecting data data from mysql

if you want to get something from "select" query you should specify the type of this field.
You can do it like this:
```
select id:int, name:sql.NullString, md5(login) as md5login:string from users
```

Program will generate structure with name like function name + "Struct"  
If our function name was "getUserId" than program will generate this structure
```
type getUserId struct {
	id int
	name sql.NullString
	md5login string
}
```

You must always use alias if your field name is not alphanumeric

Same goes for "returning" in db like postgres

### Arguments
To use arguments in sql you should use syntax of arguments:
```
delete from users where id=$id:int or login=$login:string
```

**Argument always begins with a $, than goes an identifier, then :, and then a type of argument.**

In this example we will get generated function like this:

```
function deleteUserBySomething(tx *sql.Tx, id int, login string) (sql.Result, error)
```

**If you need to use one argument more than once - you can omit the type in other usages**

### Bulks
gcgsql supports two kinds of bulks

#### []type
can be used in bulks with a comma separated values:
```
select id:int, name:string from users where id in ($ids:[]int)
```
generated function signature will be like this:
```
function getUsersByIds(tx *sql.Tx, ids []int) ([]getUsersByIdsStruct, error)
```

#### repeaters
If you need bulk insert query - you can use this syntax
```
insert into user (id,name) values $namesAndLogins#($id:int, $name:string)#
```
Program will take ($id:int, $name:string) as repeatable part and will use comma as a separator 
It will generate a structure
```
type namesAndLoginsStruct struct {
    id int
    name string
}
```
And generate a function like
```
function addUsers(tx *sql.Tx, namesAndLogins []namesAndLoginsStruct) ([]getUsersByIdsStruct, error)
```

If you need other separator you should use "|" symbol before the closing "\#"    
Everything between | and \# will be used as a separator
```
delete from users where $nameAndLastname#(name=$name:string and lastname=$lastname:string)| or # 
```

So if you need an empty separator you can use |\#

If you already have some structure and you don't want a new one you can do it by specifying the type before repeater symbol
```
delete from users where $nameAndLastname:mytype#(name=$name and lastname=$lastname)| or # 
```

when you use your own structure - you must omit the type inside the repeatable part  